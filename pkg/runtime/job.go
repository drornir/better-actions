package runtime

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/log"
	"github.com/drornir/better-actions/pkg/shell"
	"github.com/drornir/better-actions/pkg/yamls"
)

func RunJob(ctx context.Context, jobName string, job *yamls.Job) error {
	oopser := oops.FromContext(ctx).With("jobName", jobName)
	logger := log.FromContext(ctx).With("jobName", jobName)

	logger.D(ctx, "running job")

	scriptsDirPath, err := os.MkdirTemp(os.TempDir(), "bact-job-"+jobName+"-")
	if err != nil {
		return oopser.Wrapf(err, "creating scripts directory")
	}
	defer os.RemoveAll(scriptsDirPath)
	scriptDirRoot, err := os.OpenRoot(scriptsDirPath)
	if err != nil {
		return oopser.Wrapf(err, "opening scripts directory")
	}
	defer scriptDirRoot.Close()

	for i, step := range job.Steps {
		ctxkv := []any{
			"stepIndex", i,
			"step.name", step.Name,
			"step.ID", step.ID,
			"step.shell", step.Shell,
			"step.shellCommand", step.ShellCommand(),
			"step.run", step.Run,
		}
		oopser := oopser.With(ctxkv...)
		logger := logger.With(ctxkv...)
		logger.D(ctx, "running step")

		scriptName := fmt.Sprintf("%d_%s", i, step.ID)
		if err := scriptDirRoot.WriteFile(scriptName, []byte(step.Run), 0o777); err != nil {
			return oopser.Wrapf(err, "writing script file")
		}

		shellCommand := strings.Replace(step.ShellCommand(), "{0}", path.Join(scriptsDirPath, scriptName), -1)

		binArgs, err := shellquote.Split(shellCommand)
		if err != nil {
			return oopser.With("step.shellCommand", shellCommand).Wrapf(err, "parsing shell command")
		}
		bin, args := binArgs[0], binArgs[1:]
		sh, err := shell.NewShell(bin, args...)
		if err != nil {
			return oopser.With("bin", bin).With("args", args).Wrapf(err, "initializing shell")
		}
		cmd := sh.NewCommand(ctx, shell.CommandOpts{
			ExtraEnv: nil,
			Dir:      "",
			StdOut:   os.Stdout,
			StdErr:   os.Stderr,
		})

		logger.D(ctx, "running command", "command.path", cmd.Path, "command.args", cmd.Args)
		if err := cmd.Run(); err != nil {
			return oopser.With("command.path", cmd.Path).With("command.args", cmd.Args).Wrapf(err, "running command")
		}
	}

	return nil
}
