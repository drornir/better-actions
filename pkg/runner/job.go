package runner

import (
	"context"
	"os"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/log"
	"github.com/drornir/better-actions/pkg/yamls"
)

func RunJob(ctx context.Context, jobName string, job *yamls.Job) error {
	oopser := oops.FromContext(ctx).With("jobName", jobName)
	logger := log.FromContext(ctx).With("jobName", jobName)

	logger.D(ctx, "running job")

	scriptsDirPath, err := os.MkdirTemp(os.TempDir(), "bact-job-"+jobName+"-")
	if err != nil {
		return oopser.Wrapf(err, "creating scripts directory")
	}
	defer os.RemoveAll(scriptsDirPath)
	scriptDirRoot, err := os.OpenRoot(scriptsDirPath)
	if err != nil {
		return oopser.Wrapf(err, "opening scripts directory")
	}
	defer scriptDirRoot.Close()

	for i, step := range job.Steps {
		stepContext := &StepContext{
			IndexInJob:     i,
			TempScriptsDir: scriptDirRoot,
		}
		ctxkv := []any{
			"stepIndex", i,
			"step.name", step.Name,
			"step.ID", step.ID,
			// "step.shell", step.Shell,
			// "step.shellCommand", step.ShellCommand(),
			// "step.run", step.Run,
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
			res, err := sr.Exec(ctx)
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
