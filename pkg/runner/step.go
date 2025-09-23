package runner

import (
	"fmt"
	"io"
	"os"

	"github.com/drornir/better-actions/pkg/yamls"
)

type StepResult struct{}

type StepContext struct {
	Console    io.Writer
	IndexInJob int
	WorkingDir *os.Root
	Env        map[string]string
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
