package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/drornir/factor3/pkg/factor3"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/drornir/better-actions/config"
	"github.com/drornir/better-actions/log"
)

var (
	rootConfig config.Config

	rootCmd = &cobra.Command{
		Use:   "bact [global flags] [command]",
		Short: "bact is the cli entrypoint of the better-actions",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			config.SetGlobalConfig(rootConfig)
			log.SetGlobal(log.New(log.MakeSLogger(log.LoggerOptions{
				Level:  config.GetConfig().Log.Level,
				Writer: os.Stderr,
				Format: config.GetConfig().Log.Format,
			})))

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	// viperInstance := viper.NewWithOptions(viper.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	// 	Level: slog.LevelDebug,
	// }))))
	// wd, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }

	prepflag := pflag.NewFlagSet("pre", pflag.ContinueOnError)
	prepflag.StringP("config", "c", "bact.yaml", "path to config file")
	prepflag.Usage = func() {}
	prepflag.ParseErrorsWhitelist.UnknownFlags = true
	if err := prepflag.Parse(os.Args[1:]); err != nil {
		var notExistErr *pflag.NotExistError
		if !errors.Is(err, pflag.ErrHelp) && !errors.As(err, &notExistErr) {
			panic(err)
		}
	}
	cfgFilePath, err := prepflag.GetString("config")
	if err != nil {
		panic(err)
	}

	viperInstance := viper.New()
	// Setting up viper with options that fit factor3
	if err = factor3.InitializeViper(factor3.InitArgs{
		Viper:       viperInstance,
		ProgramName: "bact",
		CfgFilePath: cfgFilePath,
	}); err != nil {
		var fsperr *fs.PathError
		if !errors.Is(err, viper.ConfigFileNotFoundError{}) || !errors.As(err, &fsperr) {
			panic(err)
		}
	}

	pflags := rootCmd.PersistentFlags()
	pflags.AddFlagSet(prepflag)
	// Using Bind() we create Loader that populates the config when called
	// It also registers the flags in your pflag.FlagSet
	loader, err := factor3.Bind(&rootConfig, viperInstance, pflags)
	cobra.CheckErr(err)

	// we need to let cobra parse to commandline flags before calling Load(), so we put it in cobra.OnInitialize()
	cobra.OnInitialize(func() {
		err := loader.Load()
		cobra.CheckErr(err)
		// Advanced: You can call Load() multiple times, for example in reaction to changes to the config file.
		viperInstance.OnConfigChange(func(in fsnotify.Event) {
			if err := loader.Load(); err != nil {
				fmt.Println("error reloading config on viper.OnConfigChange")
				return
			}
			config.SetGlobalConfig(rootConfig)
		})
	})

	pflags.VisitAll(func(f *pflag.Flag) {
		if f.Name == "log-level" {
			f.Shorthand = "l"
		}
	})

	rootCmd.SetErrPrefix("better-actions error:")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
