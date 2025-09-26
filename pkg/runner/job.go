package runner

import (
	"context"
	"io"
	"maps"
	"os"
	"path"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/defers"
	"github.com/drornir/better-actions/pkg/log"
	"github.com/drornir/better-actions/pkg/yamls"
)

type Job struct {
	Name      string
	Console   io.Writer
	Config    *yamls.Job
	RunnerEnv map[string]string

	jobFilesRoot *os.Root
	stepsEnv     map[string]string
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
		ctxkv := []any{
			"stepIndex", i,
			"step.name", step.Name,
			"step.ID", step.ID,
		}
		oopser := oopser.With(ctxkv...)
		logger := logger.With(ctxkv...)
		logger.D(ctx, "running step")

		stepContext, err := j.newStepContext(ctx, i, step)
		if err != nil {
			return oopser.Wrapf(err, "creating step context")
		}

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

		if err := j.processStepEnvFile(ctx, stepContext); err != nil {
			return oopser.Wrapf(err, "processing github env file")
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

	j.stepsEnv = make(map[string]string)

	jobFilesPath, err := os.MkdirTemp(os.TempDir(), "bact-job-"+jobName+"-")
	if err != nil {
		return cleanup.Noop, oopser.Wrapf(err, "creating job files directory")
	}
	cleanup.Add(func() { os.RemoveAll(jobFilesPath) })
	jobFilesRoot, err := os.OpenRoot(jobFilesPath)
	if err != nil {
		cleanup.Run()
		return cleanup.Noop, oopser.Wrapf(err, "opening job files directory")
	}
	cleanup.Add(func() { jobFilesRoot.Close() })

	j.jobFilesRoot = jobFilesRoot

	return cleanup.Run, nil
}

func (j *Job) newStepContext(ctx context.Context, indexInJob int, step *yamls.Step) (*StepContext, error) {
	oopser := oops.FromContext(ctx)
	scriptID := scriptID(indexInJob, step)
	err := j.jobFilesRoot.MkdirAll(scriptID, 0o755)
	if err != nil {
		return nil, oopser.Wrapf(err, "creating step directory")
	}
	wd, err := os.OpenRoot(path.Join(j.jobFilesRoot.Name(), scriptID))
	if err != nil {
		return nil, oopser.Wrapf(err, "opening step directory")
	}

	env := maps.Clone(j.RunnerEnv)
	if env == nil {
		env = make(map[string]string)
	}
	maps.Copy(env, j.stepsEnv)

	for _, e := range []WFCommandEnvFile{
		GithubOutput,
		GithubState,
		GithubPath,
		GithubEnv,
		GithubStepSummary,
	} {
		f, err := wd.Create(e.FileName())
		if err != nil {
			return nil, oopser.Wrapf(err, "creating file %s", e.FileName())
		}
		f.Close()
		p := path.Join(wd.Name(), e.FileName())
		env[e.EnvVarName()] = p
	}

	return &StepContext{
		Console:    j.Console,
		IndexInJob: indexInJob,
		WorkingDir: wd,
		Env:        env,
	}, nil
}

func (j *Job) processStepEnvFile(
	ctx context.Context,
	stepCtx *StepContext,
) error {
	oopser := oops.FromContext(ctx)
	logger := log.FromContext(ctx)
	envFile, ok := stepCtx.Env[GithubEnv.EnvVarName()]
	if !ok || envFile == "" {
		return nil
	}

	updates, err := parseEnvFile(envFile)
	if err != nil {
		return oopser.With("githubEnvFile", envFile).Wrapf(err, "parsing env file")
	}
	if len(updates) == 0 {
		return nil
	}

	if j.stepsEnv == nil {
		j.stepsEnv = make(map[string]string)
	}

	for key, value := range updates {
		j.stepsEnv[key] = value
		logger.D(ctx, "applied env from github env file", "env.name", key)
	}

	return nil
}
