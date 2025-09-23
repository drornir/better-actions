package runner

import (
	"context"

	"github.com/drornir/better-actions/pkg/yamls"
)

func (r *Runner) RunWorkflow(ctx context.Context, wf *yamls.Workflow) error {
	jobs := wf.Jobs

	for jobName, job := range jobs {
		err := r.RunJob(ctx, jobName, job)
		if err != nil {
			return err
		}
	}
	return nil
}
