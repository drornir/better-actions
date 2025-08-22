package main

import (
	"encoding/json"
	"os"

	"github.com/drornir/better-actions/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configListCmd = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(config.GetConfig())
	},
}
