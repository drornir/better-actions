package shell

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/samber/oops"
)

type Shell struct {
	bin  string
	args []string
}

func NewShell(bin string, args ...string) (*Shell, error) {
	if bin == "" {
		return nil, oops.Errorf("path to shell binary was not specified")
	}
	bashPath, err := exec.LookPath(bin)
	if err != nil {
		return nil, oops.With("lookedPath", bin).Wrapf(err, "can't find executable")
	}

	return &Shell{
		bin:  bashPath,
		args: args,
	}, nil
}

type CommandOpts struct {
	Args     []string
	ExtraEnv map[string]string
	Dir      string
	StdOut   io.Writer
	StdErr   io.Writer
}

func (s *Shell) NewCommand(ctx context.Context, opts CommandOpts) *exec.Cmd {
	args := append([]string(nil), s.args...)
	args = append(args, opts.Args...)
	cmd := exec.CommandContext(ctx, s.bin, args...)
	cmd.Stdout = opts.StdOut
	cmd.Stderr = opts.StdErr
	cmd.Env = os.Environ() // TODO remove

	cmd.Dir = opts.Dir
	if opts.ExtraEnv != nil {
		extraEnv := make([]string, 0, len(opts.ExtraEnv))
		for k, v := range opts.ExtraEnv {
			extraEnv = append(extraEnv, k+"="+v)
		}
		cmd.Env = append(cmd.Env, extraEnv...)
	}
	return cmd
}
