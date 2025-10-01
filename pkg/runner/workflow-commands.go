package runner

// https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands

//go:generate go tool go-enum

import (
	"fmt"
	"strings"
)

type WFCommandEnvFile string

const (
	GithubOutput      WFCommandEnvFile = "output"
	GithubState       WFCommandEnvFile = "state"
	GithubPath        WFCommandEnvFile = "path"
	GithubEnv         WFCommandEnvFile = "env"
	GithubStepSummary WFCommandEnvFile = "step_summary"
)

func (f WFCommandEnvFile) FileName() string {
	return fmt.Sprintf("%s.txt", string(f))
}

func (f WFCommandEnvFile) EnvVarName() string {
	return fmt.Sprintf("GITHUB_%s", strings.ToUpper(string(f)))
}

type ParsedWorkflowCommand struct {
	Command WorkflowCommandName
	Props   map[string]string
	Data    string
}

// ENUM(debug)
type WorkflowCommandName string

// type TextCommandsReader struct {
// 	incoming     *bytes.Buffer
// 	downstream   io.Writer
// 	commandsChan chan UnparsedCommand

// 	err     error
// 	errLock sync.Mutex
// 	scanner *bufio.Scanner
// }

// type UnparsedCommand struct {
// 	command string
// 	args    []string
// 	opts    map[string]string
// }

// func NewTextCommandsReader(downstream io.Writer) *TextCommandsReader {
// 	var incoming bytes.Buffer
// 	scanner := bufio.NewScanner(&incoming)

// 	return &TextCommandsReader{
// 		incoming:     &incoming,
// 		downstream:   downstream,
// 		commandsChan: make(chan UnparsedCommand, 0),

// 		scanner: scanner,
// 	}
// }

// func (r *TextCommandsReader) Start(ctx context.Context) (stopper func()) {
// 	oopser := oops.FromContext(ctx)
// 	logger := log.FromContext(ctx)

// 	linesChan := make(chan string, 128)

// 	stopChan := make(chan struct{}, 1)
// 	stopper = func() {
// 		logger.D(ctx, "stopping text commands reader")
// 		stopChan <- struct{}{}
// 	}

// 	go func() {
// 		defer close(r.commandsChan)
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				r.errLock.Lock()
// 				r.err = oopser.Wrapf(ctx.Err(), "context was done")
// 				r.errLock.Unlock()
// 				return
// 			case <-stopChan:
// 				return
// 			case line := <-linesChan:
// 				logger.D(ctx, "read line", "lineStart", line[:min(len(line), 80)])
// 				unparsed, ok := r.tryParseCommand(line)
// 				if ok {
// 					r.commandsChan <- unparsed
// 				}
// 				_, err := r.downstream.Write([]byte(line))
// 				if err != nil {
// 					r.errLock.Lock()
// 					r.err = oopser.Wrapf(err, "error writing line")
// 					r.errLock.Unlock()
// 				}

// 			default:
// 				line, err := r.incoming.ReadString('\n')
// 				if err != nil {
// 					if err != io.EOF {
// 						r.errLock.Lock()
// 						r.err = oopser.Wrapf(err, "error reading line")
// 						r.errLock.Unlock()
// 					} else {
// 						logger.D(ctx, "EOF reached")
// 					}
// 					linesChan <- line
// 					close(linesChan)
// 					return
// 				}
// 			}
// 		}
// 	}()
// 	return stopper
// }

// // Write implements io.Writer.
// func (r *TextCommandsReader) Write(p []byte) (n int, err error) {
// 	r.errLock.Lock()
// 	if r.err != nil {
// 		defer r.errLock.Unlock()
// 		return 0, r.err
// 	}
// 	r.errLock.Unlock()
// 	return r.incoming.Write(p)
// }

// func (r *TextCommandsReader) CommandsChan() <-chan UnparsedCommand {
// 	return r.commandsChan
// }
