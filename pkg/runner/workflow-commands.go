package runner

// https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands

//go:generate go tool go-enum

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/drornir/better-actions/pkg/log"
)

// WorkflowCommandName is an enum for all the workflow action commands to support
// ENUM(set-env, set-output, save-state, add-mask, add-path, add-matcher, remove-matcher, debug, warning, error, notice, group, endgroup, echo)
type WorkflowCommandName string

type WFCommandEnvFile string

const (
	GithubOutput      WFCommandEnvFile = "output"
	GithubState       WFCommandEnvFile = "state"
	GithubPath        WFCommandEnvFile = "path"
	GithubEnv         WFCommandEnvFile = "env"
	GithubStepSummary WFCommandEnvFile = "step_summary"
)

func (f WFCommandEnvFile) FileName() string {
	return fmt.Sprintf("%s.txt", string(f))
}

func (f WFCommandEnvFile) EnvVarName() string {
	return fmt.Sprintf("GITHUB_%s", strings.ToUpper(string(f)))
}

type ParsedWorkflowCommand struct {
	Command WorkflowCommandName
	Props   map[string]string
	Data    string
}

func parseWorkflowCommand(ctx context.Context, line string) (ParsedWorkflowCommand, bool) {
	line = strings.TrimFunc(line, unicode.IsSpace) // also removes \r?\n from the end

	if strings.HasPrefix(line, "::") {
		return parseWorkflowCommandV2(ctx, line)
	} else if cmdStart := strings.Index(line, "##["); cmdStart >= 0 {
		line = line[cmdStart:]
		return parseWorkflowCommandV1(ctx, line)
	} else {
		return ParsedWorkflowCommand{}, false
	}
}

// parseWorkflowCommandV2 parses a command line in the format "::command key=value key=value::data". This is the
// GitHub Actions syntax that is documented and supported
func parseWorkflowCommandV2(ctx context.Context, line string) (ParsedWorkflowCommand, bool) {
	logger := log.FromContext(ctx).With("function", "parseCommandV2")

	var wfcmd ParsedWorkflowCommand
	line = strings.TrimLeft(line, " ")
	if !strings.HasPrefix(line, "::") {
		return wfcmd, false
	}
	line = line[len("::"):]
	var header, dataRaw string
	{
		headerEndIndex := strings.Index(line, "::")
		if headerEndIndex == -1 {
			return wfcmd, false
		}
		header = line[:headerEndIndex]
		dataRaw = line[headerEndIndex+len("::"):]
	}

	wfcmd.Data = unescape(escapingDataMapping, dataRaw)

	var propsStr string
	{
		firstSpaceIndex := strings.Index(header, " ")

		var commandStr string
		if firstSpaceIndex == -1 {
			commandStr = header
		} else {
			commandStr = header[:firstSpaceIndex]
			propsStr = header[firstSpaceIndex:]
		}
		c, err := ParseWorkflowCommandName(commandStr)
		if err != nil {
			logger.W(ctx, "line starts with '::"+commandStr+"' which looks like a command, but is not a known command name", "command", commandStr)
			return wfcmd, false
		}
		wfcmd.Command = c
	}

	logger = logger.With("command", wfcmd.Command)

	propsStr = strings.TrimLeft(propsStr, " ")
	for propsStr != "" {
		endOfPropIndex := strings.Index(propsStr, ",")
		if endOfPropIndex == -1 {
			endOfPropIndex = len(propsStr)
		}
		prop := propsStr[:endOfPropIndex]
		propsStr = propsStr[min(endOfPropIndex+1, len(propsStr)):]
		if prop == "" {
			continue
		}
		logger := logger.With("property", prop)
		keyValue := strings.SplitN(prop, "=", 2)
		if len(keyValue) != 2 {
			logger.W(ctx, "property '"+propsStr+"' ignored because it does not contain '='")
			continue
		}
		if keyValue[1] == "" {
			logger.W(ctx, "property '"+propsStr+"' ignored because value is empty")
			continue
		}
		if wfcmd.Props == nil {
			wfcmd.Props = make(map[string]string)
		}
		wfcmd.Props[keyValue[0]] = unescape(escapingPropertyMapping, keyValue[1])
	}

	return wfcmd, true
}

// parseWorkflowCommandV1 parses a command line in the format "##[command key=value; key=value]data". This is the
// AzureDevOps syntax that is deprecated but still supported.
func parseWorkflowCommandV1(ctx context.Context, line string) (ParsedWorkflowCommand, bool) {
	logger := log.FromContext(ctx).With("function", "parseCommandV1")

	var wfcmd ParsedWorkflowCommand

	// V1 format allows the command to appear anywhere in the line after prefix text
	// Find the start of the command
	cmdStart := strings.Index(line, "##[")
	if cmdStart == -1 {
		return wfcmd, false
	}
	line = line[cmdStart:]

	if !strings.HasPrefix(line, "##[") {
		return wfcmd, false
	}
	line = line[len("##["):]

	// Find the closing bracket
	headerEndIndex := strings.Index(line, "]")
	if headerEndIndex == -1 {
		return wfcmd, false
	}

	header := line[:headerEndIndex]
	dataRaw := line[headerEndIndex+len("]"):]

	wfcmd.Data = unescape(escapingLegacyMapping, dataRaw)

	var propsStr string
	{
		firstSpaceIndex := strings.Index(header, " ")

		var commandStr string
		if firstSpaceIndex == -1 {
			commandStr = header
		} else {
			commandStr = header[:firstSpaceIndex]
			propsStr = header[firstSpaceIndex:]
		}
		c, err := ParseWorkflowCommandName(commandStr)
		if err != nil {
			logger.W(ctx, "line starts with '##["+commandStr+"' which looks like a command, but is not a known command name", "command", commandStr)
			return wfcmd, false
		}
		wfcmd.Command = c
	}

	logger = logger.With("command", wfcmd.Command)

	propsStr = strings.TrimLeft(propsStr, " ")
	for propsStr != "" {
		endOfPropIndex := strings.Index(propsStr, ";")
		if endOfPropIndex == -1 {
			endOfPropIndex = len(propsStr)
		}
		prop := propsStr[:endOfPropIndex]
		propsStr = propsStr[min(endOfPropIndex+1, len(propsStr)):]
		if prop == "" {
			continue
		}
		logger := logger.With("property", prop)
		keyValue := strings.SplitN(prop, "=", 2)
		if len(keyValue) != 2 {
			logger.W(ctx, "property '"+prop+"' ignored because it does not contain '='")
			continue
		}
		if keyValue[1] == "" {
			logger.W(ctx, "property '"+prop+"' ignored because value is empty")
			continue
		}
		if wfcmd.Props == nil {
			wfcmd.Props = make(map[string]string)
		}
		wfcmd.Props[keyValue[0]] = unescape(escapingLegacyMapping, keyValue[1])
	}

	return wfcmd, true
}
