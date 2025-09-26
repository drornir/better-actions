package runner

import (
	"fmt"
	"io"
	"os"

	"github.com/drornir/better-actions/pkg/yamls"
)

type StepResult struct {
	Status     StepStatus
	FailReason string
}

type StepStatus string

const (
	StepStatusSucceeded           StepStatus = "Succeeded"
	StepStatusSucceededWithIssues StepStatus = "SucceededWithIssues"
	StepStatusFailed              StepStatus = "Failed"
	StepStatusCanceled            StepStatus = "Canceled"
	StepStatusSkipped             StepStatus = "Skipped"
	StepStatusAbandoned           StepStatus = "Abandoned"
)

type StepContext struct {
	Console    io.Writer
	IndexInJob int
	WorkingDir *os.Root
	Env        map[string]string
	ScriptID   string
}

func scriptID(index int, step *yamls.Step) string {
	scriptName := step.ID
	if scriptName == "" {
		scriptName = step.Name
	}
	if scriptName == "" {
		scriptName = "<unnamed>"
	}
	return fmt.Sprintf("%d_%s", index, scriptName)
}
