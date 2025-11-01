package runner

import (
	"context"
	"fmt"

	"github.com/samber/oops"
)

// JobStepOutputEvaluator is responsible for evaluating the output of a job step.
// Implements [StepOutputEvaluator]
type JobStepOutputEvaluator struct {
	job  *Job
	step *StepContext
}

func (e *JobStepOutputEvaluator) ExecuteCommand(ctx context.Context, command ParsedWorkflowCommand) error {
	oopser := oops.FromContext(ctx).With("workflow_command", command.Command)

	switch command.Command {

	case WorkflowCommandNameSetEnv:
		if !e.job.AllowUnsecureCommands() {
			return oopser.Errorf(unsupportedCommandMessageDisabled, command.Command.String())
		}
		envKey := command.Props["name"]
		if envKey == "" {
			return oopser.Errorf("environment variable name cannot be empty")
		}
		e.job.stepsEnvLock.Lock()
		defer e.job.stepsEnvLock.Unlock()
		e.job.stepsEnv[envKey] = command.Data

	case WorkflowCommandNameSetOutput:
		panic("TODO implement WorkflowCommandNameSetOutput")

	case WorkflowCommandNameSaveState:
		panic("TODO implement WorkflowCommandNameSaveState")

	case WorkflowCommandNameAddMask:
		e.job.secretsMasker.AddString(command.Data)

	case WorkflowCommandNameAddPath:
		panic("TODO implement WorkflowCommandNameAddPath")

	case WorkflowCommandNameDebug:
		panic("TODO implement WorkflowCommandNameDebug")

	case WorkflowCommandNameWarning:
		panic("TODO implement WorkflowCommandNameWarning")

	case WorkflowCommandNameError:
		panic("TODO implement WorkflowCommandNameError")

	case WorkflowCommandNameNotice:
		panic("TODO implement WorkflowCommandNameNotice")

	case WorkflowCommandNameGroup:
		panic("TODO implement WorkflowCommandNameGroup")

	case WorkflowCommandNameEndgroup:
		panic("TODO implement WorkflowCommandNameEndgroup")

	case WorkflowCommandNameEcho:
		panic("TODO implement WorkflowCommandNameEcho")

	default:
		return oopser.Code("workflow_command_not_implemented").Wrap(e.Print(ctx, fmt.Sprintf("command %s is not implemented", command.Command.String())))
	}

	return nil
}

func (e *JobStepOutputEvaluator) Print(ctx context.Context, text string) error {
	text = e.job.secretsMasker.Mask(text)

	_, err := e.step.Console.Write([]byte(text))
	if err != nil {
		oopser := oops.FromContext(ctx)
		return oopser.Wrapf(err, "writing to step console")
	}
	return nil
}
