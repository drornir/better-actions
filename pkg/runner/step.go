package runner

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

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

func stepID(index int, step *yamls.Step) string {
	scriptName := step.ID
	if scriptName == "" {
		scriptName = step.Name
	}
	if scriptName == "" {
		scriptName = "<unnamed>"
	}
	return fmt.Sprintf("%d_%s", index, sanitizeID(scriptName))
}

// sanitizeID converts an arbitrary step name/ID into a filesystem-safe
// directory name without spaces, so that GITHUB_* file paths never contain
// spaces (matching GitHub runner behavior and avoiding ambiguous redirects
// when users use unquoted redirections like >> $GITHUB_ENV).
func sanitizeID(s string) string {
	// Lowercase for stability.
	s = strings.ToLower(s)

	// Replace sequences of disallowed characters (including spaces, slashes,
	// punctuation) with a single dash. Allow only a-z, 0-9, dash, underscore, dot.
	var b strings.Builder
	b.Grow(len(s))
	lastDash := false
	for _, r := range s {
		allowed := r == '-' || r == '_' || r == '.' || unicode.IsDigit(r) || (r >= 'a' && r <= 'z')
		if allowed {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := b.String()
	out = strings.Trim(out, "-")
	if out == "" {
		return "unnamed"
	}
	return out
}
