package runner

import (
	"context"
	"io"
	"os"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/defers"
	"github.com/drornir/better-actions/pkg/log"
	"github.com/drornir/better-actions/pkg/yamls"
)

type Job struct {
	Name    string
	Console io.Writer
	Config  *yamls.Job

	scriptsDirRoot *os.Root
}

func (j *Job) Run(ctx context.Context) error {
	oopser := oops.FromContext(ctx).With("jobName", j.Name)
	logger := log.FromContext(ctx).With("jobName", j.Name)

	logger.D(ctx, "running job")

	jobCleanup, err := j.prepareJob(ctx, j.Name)
	if err != nil {
		return oopser.Wrapf(err, "preparing job")
	}
	defer jobCleanup()

	for i, step := range j.Config.Steps {
		stepContext := j.newStepContext(i)
		ctxkv := []any{
			"stepIndex", i,
			"step.name", step.Name,
			"step.ID", step.ID,
		}
		oopser := oopser.With(ctxkv...)
		logger := logger.With(ctxkv...)
		logger.D(ctx, "running step")

		var stepResult StepResult
		switch {
		case step.Run != "":
			sr := &StepRun{
				Config:  step,
				Context: stepContext,
			}
			res, err := sr.Run(ctx)
			if err != nil {
				return oopser.Wrapf(err, "executing step")
			}
			stepResult = res
		case step.Uses != "":
			// TODO
			return oopser.New("'uses' is not implemented")
		default:
			return oopser.New("step is invalid: doesn't have 'run' or 'uses'")
		}

		// TODO
		_ = stepResult
	}

	return nil
}

func (j *Job) prepareJob(
	ctx context.Context,
	jobName string,
) (_cleanup func(), _err error) {
	oopser := oops.FromContext(ctx)

	cleanup := defers.Chain{}

	scriptsDirPath, err := os.MkdirTemp(os.TempDir(), "bact-job-"+jobName+"-")
	if err != nil {
		return cleanup.Noop, oopser.Wrapf(err, "creating scripts directory")
	}
	cleanup.Add(func() { os.RemoveAll(scriptsDirPath) })
	scriptDirRoot, err := os.OpenRoot(scriptsDirPath)
	if err != nil {
		cleanup.Run()
		return cleanup.Noop, oopser.Wrapf(err, "opening scripts directory")
	}
	cleanup.Add(func() { scriptDirRoot.Close() })

	j.scriptsDirRoot = scriptDirRoot

	return cleanup.Run, nil
}

func (j *Job) newStepContext(indexInJob int) *StepContext {
	return &StepContext{
		Console:        j.Console,
		IndexInJob:     indexInJob,
		TempScriptsDir: j.scriptsDirRoot,
	}
}
