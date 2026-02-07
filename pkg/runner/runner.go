package runner

import (
	"io"
	"maps"
	"os"
	"strings"
)

type TODO any

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

func EnvFromEmptyWithBasicPath() EnvFrom {
	defaultPath := strings.Join([]string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/opt/local/bin",
		"/opt/local/sbin",
	}, string(os.PathListSeparator))

	return func() map[string]string {
		return map[string]string{
			"PATH": defaultPath,
		}
	}
}

func EnvFromMap(env map[string]string) EnvFrom {
	return func() map[string]string {
		return env
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
