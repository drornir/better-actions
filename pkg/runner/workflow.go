package runner

import (
	"context"

	"github.com/drornir/better-actions/pkg/ctxkit"
	"github.com/drornir/better-actions/pkg/yamls"
)

func (r *Runner) RunWorkflow(ctx context.Context, wf *yamls.Workflow) error {
	ctx, _, oopser := ctxkit.With(ctx, "workflow", wf.Name)
	jobs := wf.Jobs

	for jobName, job := range jobs {
		err := r.NewJob(jobName, job).Run(ctx)
		if err != nil {
			return oopser.With("job", jobName).Wrapf(err, "failed to run job")
		}
	}
	return nil
}
