package runner

// https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/concurrency"
	"github.com/drornir/better-actions/pkg/log"
)

type StepOutputEvaluator interface {
	ExecuteCommand(ctx context.Context, command ParsedWorkflowCommand)
	Print(ctx context.Context, text string)
}

type StepOutputReader struct {
	stepOutput     *bytes.Buffer
	stepOutputLock sync.Mutex
	backend        StepOutputEvaluator

	linesChan  chan string
	writesChan chan struct{}
	stopChan   chan struct{}
	doneChan   chan struct{}
	scanner    *bufio.Scanner
	stopScan   atomic.Bool
	err        error
	errLock    sync.RWMutex
}

func NewStepOutputInterpreter(backend StepOutputEvaluator) *StepOutputReader {
	var rw bytes.Buffer

	writesChan := make(chan struct{}, 128)
	stopChan := make(chan struct{}, 1)
	doneChan := make(chan struct{}, 1)

	r := &StepOutputReader{
		stepOutput: &rw,
		backend:    backend,
		linesChan:  make(chan string, 4096),
		writesChan: writesChan,
		stopChan:   stopChan,
		doneChan:   doneChan,
	}

	return r
}

func (r *StepOutputReader) Start(ctx context.Context) (stopper func()) {
	// oopser := oops.FromContext(ctx)
	logger := log.FromContext(ctx)

	stopper = func() {
		logger.D(ctx, "stopping step output reader")
		r.stopScan.Store(true)
		r.stopChan <- struct{}{}
	}

	go r.readLines(ctx)

	go r.processLines(ctx)

	return stopper
}

// Write implements io.Writer so it can read the step output
func (r *StepOutputReader) Write(p []byte) (n int, err error) {
	r.stepOutputLock.Lock()
	defer r.stepOutputLock.Unlock()
	return r.stepOutput.Write(p)
}

func (r *StepOutputReader) readLines(ctx context.Context) {
	ctxkv := []any{"output_reader_worker", "read_step_output"}
	oopser := oops.FromContext(ctx).With(ctxkv...)
	logger := log.FromContext(ctx).With(ctxkv...)

	defer func() { close(r.linesChan) }()

	var lineBuf bytes.Buffer
	// the reason for this function is to do 'defer r.stepOutputLock.Unlock()' but not keeping the lock when waiting for more input
	read := func() (line string, readNewBytes, gotEOF bool, err error) {
		r.stepOutputLock.Lock()
		defer r.stepOutputLock.Unlock()

		l, err := r.stepOutput.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", false, false, oopser.Wrapf(err, "unexpected error reading from input")
		}
		if l != "" {
			if _, err := lineBuf.WriteString(l); err != nil {
				return "", false, false, oopser.Wrapf(err, "unexpected error writing to internal line buffer")
			}
		}
		if errors.Is(err, io.EOF) {
			if r.stopScan.Load() {
				fullLine := lineBuf.String()
				lineBuf.Reset()
				return fullLine, len(l) > 0, true, nil
			}
			return "", len(l) > 0, true, nil
		}

		fullLine := lineBuf.String()
		lineBuf.Reset()
		return fullLine, len(l) > 0, false, nil
	}

	const (
		initialBackoff = time.Millisecond
		maxBackoff     = time.Second * 5
	)

	backoff := initialBackoff
	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, readNewBytes, gotEOF, err := read()
			if err != nil {
				r.setErr(err)
				return
			}
			if readNewBytes {
				backoff = initialBackoff
			}
			if gotEOF {
				if r.stopScan.Load() {
					if len(line) > 0 {
						r.linesChan <- line
					}
					return
				}
				backoff = min(backoff*2, maxBackoff)
				logger.With("backoff_amount", backoff.String()).
					D(ctx, "EOF backoff: waiting %s for more input", backoff)
				select {
				case <-ctx.Done():
				case <-time.After(backoff):
				}
				continue
			}
			r.linesChan <- line
		}
	}
}

func (r *StepOutputReader) processLines(ctx context.Context) {
	ctxkv := []any{"output_reader_worker", "process_lines"}
	ctx = oops.WithBuilder(ctx, oops.FromContext(ctx).With(ctxkv...))
	ctx = log.FromContext(ctx).With(ctxkv...).ContextWithLogger(ctx)

	for line := range concurrency.ClosedOrDone(r.linesChan, ctx) {
		wfcmd, ok := parseWorkflowCommand(ctx, line)
		if ok {
			r.backend.ExecuteCommand(ctx, wfcmd)
			continue
		}
		r.backend.Print(ctx, line)
	}
}

func (r *StepOutputReader) Err() error {
	r.errLock.RLock()
	defer r.errLock.RUnlock()
	return r.err
}

func (r *StepOutputReader) setErr(err error) {
	r.errLock.Lock()
	defer r.errLock.Unlock()
	if r.err == nil {
		r.err = err
	}
}
