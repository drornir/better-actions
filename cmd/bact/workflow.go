package main

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"

	"github.com/samber/oops"
	"github.com/spf13/cobra"

	"github.com/drornir/better-actions/pkg/runner"
	"github.com/drornir/better-actions/pkg/types"
	"github.com/drornir/better-actions/pkg/yamls"
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

// runWorkflowParams are flags that capture the standard data like github, inputs, secrets, vars
// all values a re expted to be jsons.
var runWorkflowParams struct {
	github  string
	env     string
	inputs  string
	secrets string
	vars    string
	runner  string
}

func init() {
	// Add workflow command to root
	rootCmd.AddCommand(workflowCmd)

	// Add run subcommand to workflow
	workflowCmd.AddCommand(workflowRunCmd)

	// Add flags to run command
	workflowRunCmd.Flags().StringVarP(&workflowFile, "file", "f", "", "Path to the workflow file")
	workflowRunCmd.MarkFlagRequired("file")

	workflowRunCmd.Flags().StringVar(&runWorkflowParams.github, "github", "", "GitHub data")
	workflowRunCmd.Flags().StringVar(&runWorkflowParams.env, "env", "", "Environment data")
	workflowRunCmd.Flags().StringVar(&runWorkflowParams.inputs, "inputs", "", "Inputs data")
	workflowRunCmd.Flags().StringVar(&runWorkflowParams.secrets, "secrets", "", "Secrets data")
	workflowRunCmd.Flags().StringVar(&runWorkflowParams.vars, "vars", "", "Variables data")
	workflowRunCmd.Flags().StringVar(&runWorkflowParams.runner, "runner", "", "Runner data")
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

	var wfContext types.WorkflowContexts
	{
		unmarshals := []struct {
			Name    string
			Value   string
			Pointer any
		}{
			{"github", runWorkflowParams.github, &wfContext.GitHub},
			{"env", runWorkflowParams.env, &wfContext.Env},
			{"inputs", runWorkflowParams.inputs, &wfContext.Inputs},
			{"secrets", runWorkflowParams.secrets, &wfContext.Secrets},
			{"vars", runWorkflowParams.vars, &wfContext.Vars},
			{"runner", runWorkflowParams.runner, &wfContext.Runner},
		}
		for _, part := range unmarshals {
			if part.Value == "" {
				continue
			}
			err := json.Unmarshal([]byte(part.Value), part.Pointer)
			if err != nil {
				return oops.Errorf("failed to unmarshal %s data: %w", part.Name, err)
			}
		}
	}

	if err := executeWorkflowFile(ctx, absPath, &wfContext); err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	fmt.Println("Workflow completed successfully")
	return nil
}

func executeWorkflowFile(ctx context.Context, filePath string, wfContext *types.WorkflowContexts) error {
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

	rnr := runner.New(os.Stdout, runner.EnvFromChain(runner.EnvFromOS(), runner.EnvFromMap(wfContext.Env)))

	_, err2 := rnr.RunWorkflow(ctx, wf, wfContext)
	return err2
}
