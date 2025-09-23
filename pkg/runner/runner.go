package runner

import (
	"io"
	"maps"
	"os"
	"strings"

	"github.com/drornir/better-actions/pkg/yamls"
)

type Runner struct {
	Console io.Writer
	Env     map[string]string
}

func New(console io.Writer, envFrom EnvFrom) *Runner {
	return &Runner{
		Console: console,
		Env:     envFrom(),
	}
}

func (r *Runner) NewJob(name string, yaml *yamls.Job) *Job {
	return &Job{
		Name:      name,
		Console:   r.Console,
		Config:    yaml,
		RunnerEnv: r.Env,
	}
}

type EnvFrom func() map[string]string

func EnvFromEnviron(environ []string) EnvFrom {
	env := make(map[string]string)
	for _, pair := range environ {
		parts := strings.SplitN(pair, "=", 2)
		env[parts[0]] = parts[1]
	}
	return func() map[string]string {
		return env
	}
}

func EnvFromOS() EnvFrom {
	return EnvFromEnviron(os.Environ())
}

func EnvFromEmpty() EnvFrom {
	return func() map[string]string {
		return nil
	}
}

func EnvFromChain(envs ...EnvFrom) EnvFrom {
	env := make(map[string]string)
	for _, from := range envs {
		maps.Copy(env, from())
	}
	return func() map[string]string {
		return env
	}
}
