package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/kballard/go-shellquote"

	"github.com/drornir/better-actions/pkg/ctxkit"
	"github.com/drornir/better-actions/pkg/shell"
	"github.com/drornir/better-actions/pkg/yamls"
)

type StepRun struct {
	Config  *yamls.Step
	Context *StepContext
}

func (s *StepRun) Run(ctx context.Context, writeTo io.Writer) (StepResult, error) {
	step := s.Config
	wd := s.Context.WorkingDir

	ctx, logger, oopser := ctxkit.With(ctx,
		"step.shell", step.Shell,
		"step.shellCommand", step.ShellCommand(),
		"step.run", step.Run)

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

	workDir := s.Config.WorkingDirectory
	if workDir == "" {
		workDir = s.Context.WorkspaceDir
	} else if filepath.IsAbs(workDir) {
		return StepResult{}, oopser.Errorf("absolute paths are not allowed in working-directory: %s", workDir)
	} else {
		workDir = filepath.Join(s.Context.WorkspaceDir, workDir)
	}

	cmd := sh.NewCommand(ctx, shell.CommandOpts{
		ExtraEnv: s.Context.Env,
		Dir:      workDir,
		StdOut:   writeTo,
		StdErr:   writeTo,
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

	return StepResult{
		Status: StepStatusSucceeded,
	}, nil
}
