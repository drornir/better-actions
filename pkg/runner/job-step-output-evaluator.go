package runner

import (
	"context"

	"github.com/samber/oops"
)

// JobStepOutputEvaluator is responsible for evaluating the output of a job step.
// Implements [StepOutputEvaluator]
type JobStepOutputEvaluator struct {
	job  *Job
	step *StepContext
}

func (e *JobStepOutputEvaluator) ExecuteCommand(ctx context.Context, command ParsedWorkflowCommand) error {
	// Implementation of ExecuteCommand
	return nil
}

func (e *JobStepOutputEvaluator) Print(ctx context.Context, text string) error {
	// TODO remove sensitive data
	//
	_, err := e.step.Console.Write([]byte(text))
	if err != nil {
		oopser := oops.FromContext(ctx)
		return oopser.Wrapf(err, "writing to step console")
	}
	return nil
}
