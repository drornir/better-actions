package runtime

import (
	"context"

	"github.com/drornir/better-actions/workflow"
)

func RunWorkflow(ctx context.Context, wf *workflow.Workflow) error {
	jobs := wf.YAML.Jobs

	for jobName, job := range jobs {
		err := RunJob(ctx, jobName, job)
		if err != nil {
			return err
		}
	}
	return nil
}
