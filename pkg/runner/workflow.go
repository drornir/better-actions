package runner

import (
	"context"

	"github.com/drornir/better-actions/pkg/ctxkit"
	"github.com/drornir/better-actions/pkg/yamls"
)

func (r *Runner) RunWorkflow(ctx context.Context, wf *yamls.Workflow) (*WorkflowState, error) {
	ctx, _, oopser := ctxkit.With(ctx, "workflow", wf.Name)
	jobs := wf.Jobs

	wfState := &WorkflowState{
		Name: wf.Name,
		Jobs: make(map[string]*Job, len(jobs)),
	}

	for jobName, job := range jobs {
		j := r.NewJob(jobName, job)
		wfState.Jobs[jobName] = j
		err := j.Run(ctx)
		if err != nil {
			return wfState, oopser.With("job", jobName).Wrapf(err, "failed to run job")
		}
	}
	return wfState, nil
}

type WorkflowState struct {
	Name string
	Jobs map[string]*Job
}
