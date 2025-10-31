package runner

import (
	"context"
	"strings"

	"github.com/samber/oops"
)

// JobStepOutputEvaluator is responsible for evaluating the output of a job step.
// Implements [StepOutputEvaluator]
type JobStepOutputEvaluator struct {
	job  *Job
	step *StepContext
}

func (e *JobStepOutputEvaluator) ExecuteCommand(ctx context.Context, command ParsedWorkflowCommand) error {
	return nil
}

func (e *JobStepOutputEvaluator) Print(ctx context.Context, text string) error {
	for _, sens := range e.job.sensitiveStrings {
		text = strings.ReplaceAll(text, sens, "***")
	}
	for _, senseexp := range e.job.sensitiveRegexes {
		text = senseexp.ReplaceAllLiteralString(text, "***")
	}

	_, err := e.step.Console.Write([]byte(text))
	if err != nil {
		oopser := oops.FromContext(ctx)
		return oopser.Wrapf(err, "writing to step console")
	}
	return nil
}
