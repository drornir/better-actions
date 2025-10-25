package runner

import (
	"context"
	"io"
	"maps"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

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

	jobFilesRoot  *os.Root
	stepsEnv      map[string]string
	stepsPath     []string
	stepOutputs   map[string]map[string]string
	stepStates    map[string]map[string]string
	stepSummaries map[string]string

	sensitiveStrings []string
	sensitiveRegexes []regexp.Regexp
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
			"step.ID", makeStepID(i, step),
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

		if stepResult.Status == StepStatusFailed {
			// TODO this step failed but I need to check the conditions on the next steps and possibly move on
			return oopser.
				Wrapf(oops.New(stepResult.FailReason), "step failed")
		}

		if err := j.loadWFCmdFilesAfterStep(ctx, stepContext); err != nil {
			return oopser.Wrapf(err, "processing workflow command files")
		}
		// TODO
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
	j.stepsPath = nil
	j.stepOutputs = make(map[string]map[string]string)
	j.stepStates = make(map[string]map[string]string)
	j.stepSummaries = make(map[string]string)

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
	stpID := makeStepID(indexInJob, step)
	err := j.jobFilesRoot.MkdirAll(stpID, 0o755)
	if err != nil {
		return nil, oopser.Wrapf(err, "creating step directory")
	}
	wd, err := os.OpenRoot(path.Join(j.jobFilesRoot.Name(), stpID))
	if err != nil {
		return nil, oopser.Wrapf(err, "opening step directory")
	}

	env := maps.Clone(j.RunnerEnv)
	if env == nil {
		env = make(map[string]string)
	}
	maps.Copy(env, j.stepsEnv)
	j.applyPrependPath(ctx, env)

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
		StepID:     stpID,
		Console:    j.Console,
		IndexInJob: indexInJob,
		WorkingDir: wd,
		Env:        env,
	}, nil
}

func (j *Job) applyPrependPath(ctx context.Context, env map[string]string) {
	logger := log.FromContext(ctx)
	if len(j.stepsPath) == 0 {
		return
	}

	originalPath, ok := env["PATH"]
	if !ok {
		originalPath = env["Path"]
	}

	entries := make([]string, 0, len(j.stepsPath)+1)
	for _, pathEntry := range slices.Backward(j.stepsPath) {
		entries = append(entries, pathEntry)
	}
	if originalPath != "" {
		entries = append(entries, originalPath)
	}

	newPath := strings.Join(entries, string(os.PathListSeparator))
	env["PATH"] = newPath
	if _, hasPath := env["Path"]; hasPath {
		env["Path"] = newPath
	}
	logger.D(ctx, "applyPrependPath",
		"fromSteps", strings.Join(j.stepsPath, string(os.PathListSeparator)),
		"original", originalPath,
		"newPath", newPath)
}

func (j *Job) loadWFCmdFilesAfterStep(
	ctx context.Context,
	stepCtx *StepContext,
) error {
	oopser := oops.FromContext(ctx).With("job_life_cycle", "loadWFCmdFilesAfterStep")

	if err := j.loadEnvWfCmdFile(ctx, stepCtx); err != nil {
		return oopser.Wrapf(err, "processing env command file")
	}
	if err := j.loadPathWfCmdFile(ctx, stepCtx); err != nil {
		return oopser.Wrapf(err, "processing path command file")
	}
	if err := j.loadOutputWfCmdFile(ctx, stepCtx); err != nil {
		return oopser.Wrapf(err, "processing output command file")
	}
	if err := j.loadStateWfCmdFile(ctx, stepCtx); err != nil {
		return oopser.Wrapf(err, "processing state command file")
	}
	if err := j.loadStepSummaryWfCmdFile(ctx, stepCtx); err != nil {
		return oopser.Wrapf(err, "processing step summary command file")
	}

	return nil
}

func (j *Job) loadEnvWfCmdFile(ctx context.Context, stepCtx *StepContext) error {
	logger := log.FromContext(ctx)
	oopser := oops.FromContext(ctx)

	filePath, ok := j.commandFilePath(stepCtx, GithubEnv)
	if !ok {
		return nil
	}

	updates, err := parseCommandKeyValueFile(filePath, GithubEnv)
	if err != nil {
		return oopser.Wrapf(err, "parsing env file")
	}

	for key, value := range updates {
		j.stepsEnv[key] = value
		logger.D(ctx, "applied env from github env file", "env.name", key)
	}

	return nil
}

func (j *Job) loadPathWfCmdFile(ctx context.Context, stepCtx *StepContext) error {
	logger := log.FromContext(ctx)
	oopser := oops.FromContext(ctx)

	filePath, ok := j.commandFilePath(stepCtx, GithubPath)
	if !ok {
		return nil
	}

	entries, err := parsePathFile(filePath)
	if err != nil {
		return oopser.Wrapf(err, "parsing path file")
	}
	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		j.addPathEntry(entry)
		logger.D(ctx, "applied path from github path file", "path.entry", entry)
	}

	return nil
}

func (j *Job) loadOutputWfCmdFile(ctx context.Context, stepCtx *StepContext) error {
	logger := log.FromContext(ctx)
	oopser := oops.FromContext(ctx)

	filePath, ok := j.commandFilePath(stepCtx, GithubOutput)
	if !ok {
		return nil
	}

	updates, err := parseCommandKeyValueFile(filePath, GithubOutput)
	if err != nil {
		return oopser.Wrapf(err, "parsing output file")
	}
	if len(updates) == 0 {
		return nil
	}

	stepKey := stepCtx.StepID
	output := j.stepOutputs[stepKey]
	if output == nil {
		output = make(map[string]string)
		j.stepOutputs[stepKey] = output
	}

	for key, value := range updates {
		output[key] = value
		logger.D(ctx, "captured step output", "step.scriptID", stepKey, "output.name", key)
	}

	return nil
}

func (j *Job) loadStateWfCmdFile(ctx context.Context, stepCtx *StepContext) error {
	logger := log.FromContext(ctx)
	oopser := oops.FromContext(ctx)

	filePath, ok := j.commandFilePath(stepCtx, GithubState)
	if !ok {
		return nil
	}

	updates, err := parseCommandKeyValueFile(filePath, GithubState)
	if err != nil {
		return oopser.Wrapf(err, "parsing state file")
	}
	if len(updates) == 0 {
		return nil
	}

	stepKey := stepCtx.StepID
	state := j.stepStates[stepKey]
	if state == nil {
		state = make(map[string]string)
		j.stepStates[stepKey] = state
	}

	for key, value := range updates {
		state[key] = value
		logger.D(ctx, "captured step state", "step.scriptID", stepKey, "state.name", key)
	}

	return nil
}

func (j *Job) loadStepSummaryWfCmdFile(ctx context.Context, stepCtx *StepContext) error {
	logger := log.FromContext(ctx)
	oopser := oops.FromContext(ctx)

	filePath, ok := j.commandFilePath(stepCtx, GithubStepSummary)
	if !ok {
		return nil
	}

	summary, err := readStepSummary(filePath)
	if err != nil {
		return oopser.Wrapf(err, "reading step summary")
	}

	j.stepSummaries[stepCtx.StepID] = summary
	logger.D(ctx, "captured step summary", "step.scriptID", stepCtx.StepID)
	return nil
}

func (j *Job) commandFilePath(stepCtx *StepContext, command WFCommandEnvFile) (string, bool) {
	if stepCtx == nil {
		return "", false
	}
	path, ok := stepCtx.Env[command.EnvVarName()]
	if !ok || path == "" {
		return "", false
	}
	return path, true
}

func (j *Job) addPathEntry(entry string) {
	if entry == "" {
		return
	}

	for i, existing := range j.stepsPath {
		if existing == entry {
			j.stepsPath = append(j.stepsPath[:i], j.stepsPath[i+1:]...)
			break
		}
	}

	j.stepsPath = append(j.stepsPath, entry)
}

func readStepSummary(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
