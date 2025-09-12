package runner

import (
	"fmt"
	"os"

	"github.com/drornir/better-actions/pkg/yamls"
)

type StepResult struct{}

type StepContext struct {
	IndexInJob     int
	TempScriptsDir *os.Root
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
