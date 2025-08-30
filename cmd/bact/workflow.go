package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/drornir/better-actions/pkg/runtime"
	"github.com/drornir/better-actions/pkg/yamls"
	"github.com/drornir/better-actions/workflow"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Workflow management commands",
	Long:  "Commands for managing and executing workflows",
}

var workflowRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a workflow",
	Long:  "Execute a workflow from a file",
	RunE:  runWorkflow,
}

var workflowFile string

func init() {
	// Add workflow command to root
	rootCmd.AddCommand(workflowCmd)

	// Add run subcommand to workflow
	workflowCmd.AddCommand(workflowRunCmd)

	// Add flags to run command
	workflowRunCmd.Flags().StringVarP(&workflowFile, "file", "f", "", "Path to the workflow file")
	workflowRunCmd.MarkFlagRequired("file")
}

func runWorkflow(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if workflowFile == "" {
		return fmt.Errorf("workflow file is required")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(workflowFile)
	if err != nil {
		return fmt.Errorf("failed to resolve workflow file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("workflow file does not exist: %s", absPath)
	}

	fmt.Printf("Running workflow from: %s\n", absPath)

	if err := executeWorkflowFile(ctx, absPath); err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	fmt.Println("Workflow completed successfully")
	return nil
}

func executeWorkflowFile(ctx context.Context, filePath string) error {
	// Read the workflow file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read workflow file: %w", err)
	}

	fmt.Printf("Workflow file size: %d bytes\n", len(content))

	// TODO: Parse and execute the workflow
	// This is where you would implement the actual workflow parsing and execution logic
	// For now, we'll just print the file contents
	fmt.Println("Workflow content:")
	fmt.Println(string(content))

	openFile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	wf, err := yamls.ReadWorkflow(openFile, false)
	if err != nil {
		return err
	}

	return runtime.RunWorkflow(ctx, &workflow.Workflow{YAML: wf})
}
