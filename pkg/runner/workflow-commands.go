package runner

// https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands

import (
	"fmt"
	"strings"
)

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
