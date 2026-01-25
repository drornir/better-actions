package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/ctxkit"
	"github.com/drornir/better-actions/pkg/log"
)

// JobStepOutputEvaluator is responsible for evaluating the output of a job step.
// Implements [StepOutputEvaluator]
type JobStepOutputEvaluator struct {
	job  *Job
	step *StepContext
}

// ExecuteCommand runs commands that were processed via outputting string to stdout from the user's script/action
// TODO 'notice', 'warning', 'error' and 'add-matcher' are not yet implemented (as of 2025-11)
func (e *JobStepOutputEvaluator) ExecuteCommand(ctx context.Context, command ParsedWorkflowCommand) error {
	ctx, logger, oopser := ctxkit.With(ctx, "workflow_command", command.Command)

	echoIfEnabled := func() {
		if !e.step.EchoCommands {
			return
		}
		if err := e.Print(ctx, command.RawString); err != nil {
			logger.E(ctx, "error echoing command", "error", err)
		}
	}

	switch command.Command {

	case WorkflowCommandNameSetEnv:
		if !e.job.AllowUnsecureCommands() {
			return oopser.Errorf(unsupportedCommandMessageDisabled, command.Command.String())
		}
		echoIfEnabled()
		key := command.Props["name"]
		if key == "" {
			return oopser.Errorf("environment variable name cannot be empty")
		}
		entry := encodeEnvfileLikeKeyValue(key, command.Data)
		return e.job.appendToCommandFile(ctx, e.step, GithubEnv, entry)

	case WorkflowCommandNameSetOutput:
		echoIfEnabled()
		key := command.Props["name"]
		if key == "" {
			return oopser.Errorf("output variable name cannot be empty")
		}
		entry := encodeEnvfileLikeKeyValue(key, command.Data)
		return e.job.appendToCommandFile(ctx, e.step, GithubOutput, entry)

	case WorkflowCommandNameSaveState:
		echoIfEnabled()
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
		echoIfEnabled()
		if command.Data == "" {
			return oopser.Errorf("path cannot be empty")
		}
		return e.job.appendToCommandFile(ctx, e.step, GithubPath, command.Data)

	case WorkflowCommandNameAddMatcher, WorkflowCommandNameRemoveMatcher:
		logger.W(ctx, "matcher type commands are not supported yet (add-matcher, remove-matcher)")
		return nil

	case WorkflowCommandNameDebug:
		if !e.job.debugEnabled {
			return nil
		}
		cleanString := strings.ReplaceAll(command.RawString, "\r\n", "\n")
		for l := range strings.SplitSeq(cleanString, "\n") {
			if err := e.Print(ctx, fmt.Sprintf("##[debug] %s", l)); err != nil {
				return oopser.Wrap(err)
			}
		}
		return nil

	case WorkflowCommandNameNotice, WorkflowCommandNameWarning, WorkflowCommandNameError:
		return e.processIssueCommand(ctx, command)

	case WorkflowCommandNameGroup:
		echoIfEnabled()
		return oopser.Wrap(e.Print(ctx, fmt.Sprintf("##[group]%s", command.Data)))

	case WorkflowCommandNameEndgroup:
		echoIfEnabled()
		return oopser.Wrap(e.Print(ctx, fmt.Sprintf("##[endgroup]%s", command.Data)))

	case WorkflowCommandNameEcho:
		echoIfEnabled()
		switch strings.ToUpper(strings.TrimSpace(command.Data)) {
		case "ON":
			e.step.EchoCommands = true
		case "OFF":
			e.step.EchoCommands = false
		default:
			return oopser.Errorf("echo command accepts only 'on' or 'off', got '%s'", command.Command)
		}
		return nil

	default:
		logger.E(ctx, "unknown workflow command")
		return oopser.Wrap(e.Print(ctx, command.RawString))
	}
}

// processIssueCommand handles 'notice', 'warning', and 'error' commands.
// reference: `public abstract class IssueCommandExtension` in Runner.Worker/ActionCommandManager.cs:600
func (e *JobStepOutputEvaluator) processIssueCommand(ctx context.Context, command ParsedWorkflowCommand) error {
	_ = command
	logger := log.FromContext(ctx)
	logger.W(ctx, "issue type commands are not supported yet (notice, warn, error)")
	return nil
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
