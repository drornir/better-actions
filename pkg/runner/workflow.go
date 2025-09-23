package runner

import (
	"context"

	"github.com/drornir/better-actions/pkg/yamls"
)

func (r *Runner) RunWorkflow(ctx context.Context, wf *yamls.Workflow) error {
	jobs := wf.Jobs

	for jobName, job := range jobs {
		err := r.NewJob(jobName, job).Run(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
