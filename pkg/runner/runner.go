package runner

import (
	"io"

	"github.com/drornir/better-actions/pkg/yamls"
)

type Runner struct {
	Console io.Writer
}

func (r *Runner) NewJob(name string, yaml *yamls.Job) *Job {
	return &Job{
		Name:    name,
		Console: r.Console,
		Config:  yaml,
	}
}
