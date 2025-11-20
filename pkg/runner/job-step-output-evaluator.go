package runner

import (
	"context"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/ctxkit"
)

// JobStepOutputEvaluator is responsible for evaluating the output of a job step.
// Implements [StepOutputEvaluator]
type JobStepOutputEvaluator struct {
	job  *Job
	step *StepContext
}

func (e *JobStepOutputEvaluator) ExecuteCommand(ctx context.Context, command ParsedWorkflowCommand) error {
	ctx, logger, oopser := ctxkit.With(ctx, "workflow_command", command.Command)

	switch command.Command {

	case WorkflowCommandNameSetEnv:
		if !e.job.AllowUnsecureCommands() {
			return oopser.Errorf(unsupportedCommandMessageDisabled, command.Command.String())
		}
		key := command.Props["name"]
		if key == "" {
			return oopser.Errorf("environment variable name cannot be empty")
		}
		entry := encodeEnvfileLikeKeyValue(key, command.Data)
		return e.job.appendToCommandFile(ctx, e.step, GithubEnv, entry)

	case WorkflowCommandNameSetOutput:
		key := command.Props["name"]
		if key == "" {
			return oopser.Errorf("output variable name cannot be empty")
		}
		entry := encodeEnvfileLikeKeyValue(key, command.Data)
		return e.job.appendToCommandFile(ctx, e.step, GithubOutput, entry)

	case WorkflowCommandNameSaveState:
		key := command.Props["name"]
		if key == "" {
			return oopser.Errorf("state key name cannot be empty")
		}
		entry := encodeEnvfileLikeKeyValue(key, command.Data)
		return e.job.appendToCommandFile(ctx, e.step, GithubState, entry)

	case WorkflowCommandNameAddMask:
		e.job.secretsMasker.AddString(command.Data)
		return nil

	case WorkflowCommandNameAddPath:
		if command.Data == "" {
			return oopser.Errorf("path cannot be empty")
		}
		return e.job.appendToCommandFile(ctx, e.step, GithubPath, command.Data)

	case WorkflowCommandNameDebug:
		if e.job.debugEnabled {
			return oopser.Wrap(e.Print(ctx, command.RawString))
		}
		return nil

	case WorkflowCommandNameWarning:
		panic("TODO implement WorkflowCommandNameWarning")

	case WorkflowCommandNameError:
		panic("TODO implement WorkflowCommandNameError")

	case WorkflowCommandNameNotice:
		panic("TODO implement WorkflowCommandNameNotice")

	case WorkflowCommandNameGroup:
		return oopser.Wrap(e.Print(ctx, command.RawString))

	case WorkflowCommandNameEndgroup:
		return oopser.Wrap(e.Print(ctx, command.RawString))

	case WorkflowCommandNameEcho:
		panic("TODO implement WorkflowCommandNameEcho")

	default:
		logger.E(ctx, "unknown workflow command")
		return oopser.Wrap(e.Print(ctx, command.RawString))
	}
}

func (e *JobStepOutputEvaluator) Print(ctx context.Context, text string) error {
	text = e.job.secretsMasker.Mask(text)
	textb := []byte(text)
	if text == "" || text[len(text)-1] != '\n' {
		textb = append(textb, '\n')
	}

	_, err := e.step.Console.Write(textb)
	if err != nil {
		oopser := oops.FromContext(ctx)
		return oopser.Wrapf(err, "writing to step console")
	}
	return nil
}
