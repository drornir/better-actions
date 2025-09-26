package runner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/log"
	"github.com/drornir/better-actions/pkg/shell"
	"github.com/drornir/better-actions/pkg/yamls"
)

type StepRun struct {
	Config  *yamls.Step
	Context *StepContext
}

func (s *StepRun) Run(ctx context.Context) (StepResult, error) {
	oopser := oops.FromContext(ctx)
	logger := log.FromContext(ctx)

	step := s.Config
	wd := s.Context.WorkingDir

	ctxkv := []any{
		"step.shell", step.Shell,
		"step.shellCommand", step.ShellCommand(),
		"step.run", step.Run,
	}
	oopser = oopser.With(ctxkv...)
	logger = logger.With(ctxkv...)

	const scriptName = "script.sh"
	if err := wd.WriteFile(scriptName, []byte(step.Run), 0o777); err != nil {
		return StepResult{}, oopser.With("scriptFile", path.Join(wd.Name(), scriptName)).
			Wrapf(err, "writing script file")
	}

	shellCommand := strings.ReplaceAll(step.ShellCommand(), "{0}",
		shellquote.Join(path.Join(wd.Name(), scriptName)),
	)

	binArgs, err := shellquote.Split(shellCommand)
	if err != nil {
		return StepResult{}, oopser.Wrapf(err, "parsing shell command")
	}
	bin, args := binArgs[0], binArgs[1:]
	sh, err := shell.NewShell(bin, args...)
	if err != nil {
		return StepResult{}, oopser.With("step.shell.bin", bin).With("step.shell.args", args).Wrapf(err, "initializing shell")
	}
	cmd := sh.NewCommand(ctx, shell.CommandOpts{
		ExtraEnv: s.Context.Env,
		Dir:      s.Config.WorkingDirectory,
		StdOut:   s.Context.Console,
		StdErr:   s.Context.Console,
	})

	logger.D(ctx, "running command", "command.path", cmd.Path, "command.args", cmd.Args)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return StepResult{
				Status:     StepStatusFailed,
				FailReason: fmt.Sprintf("%s returned %s", shellquote.Join(cmd.Args...), exitErr.Error()),
			}, nil
		}
		return StepResult{}, oopser.With("command.path", cmd.Path).With("command.args", cmd.Args).Wrapf(err, "running command")
	}

	return StepResult{}, nil
}
