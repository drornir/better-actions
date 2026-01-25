package expr

// EvalContext represents the context available during expression evaluation.
// It includes information about the workflow, job, steps, and other metadata.
type EvalContext struct {
	// Github contains information about the workflow run and the event that triggered it.
	Github GithubContext `json:"github"`
	// Env contains environment variables set in the workflow, job, or step.
	Env map[string]string `json:"env"`
	// Job contains information about the currently running job.
	Job JobContext `json:"job"`
	// Jobs contains information about jobs in a reusable workflow.
	// This is only available in reusable workflows.
	Jobs JobsContext `json:"jobs"`
	// Steps contains information about the steps that have been run in the current job.
	Steps StepsContext `json:"steps"`
	// Runner contains information about the runner that is executing the current job.
	Runner RunnerContext `json:"runner"`
	// Secrets contains secrets available to the workflow run.
	Secrets SecretsContext `json:"secrets"`
	// Vars contains variables available to the workflow run.
	Vars map[string]string `json:"vars"`
	// Strategy contains information about the matrix execution strategy.
	Strategy StrategyContext `json:"strategy"`
	// Matrix contains information about the specific matrix combination for the current job.
	Matrix JSObject `json:"matrix"`
	// Needs contains outputs and results from jobs that the current job depends on.
	Needs map[string]NeedsContext `json:"needs"`
	// Inputs contains inputs passed to the workflow (dispatch inputs or reusable workflow inputs).
	Inputs JSObject `json:"inputs"`
}

// GithubContext is modeled after https://docs.github.com/en/actions/reference/workflows-and-actions/contexts#github-context
// It provides information about the workflow run and the event that triggered it.
type GithubContext struct {
	// Action is the name of the action currently running, or the id of a step.
	Action string `json:"action"`
	// ActionPath is the path where an action is located. Only supported in composite actions.
	ActionPath string `json:"action_path"`
	// ActionRef is the ref of the action being executed (e.g., v2).
	ActionRef string `json:"action_ref"`
	// ActionRepository is the owner and repository name of the action (e.g., actions/checkout).
	ActionRepository string `json:"action_repository"`
	// ActionStatus is the current result of the composite action.
	ActionStatus string `json:"action_status"`
	// Actor is the username of the user that triggered the initial workflow run.
	Actor string `json:"actor"`
	// ActorID is the account ID of the person or app that triggered the initial workflow run.
	ActorID string `json:"actor_id"`
	// APIURL is the URL of the GitHub REST API.
	APIURL string `json:"api_url"`
	// BaseRef is the base_ref or target branch of the pull request in a workflow run.
	BaseRef string `json:"base_ref"`
	// Env is the path on the runner to the file that sets environment variables from workflow commands.
	Env string `json:"env"`
	// Event is the full event webhook payload.
	Event JSObject `json:"event"`
	// EventName is the name of the event that triggered the workflow run.
	EventName string `json:"event_name"`
	// EventPath is the path to the file on the runner that contains the full event webhook payload.
	EventPath string `json:"event_path"`
	// GraphQLURL is the URL of the GitHub GraphQL API.
	GraphQLURL string `json:"graphql_url"`
	// HeadRef is the head_ref or source branch of the pull request in a workflow run.
	HeadRef string `json:"head_ref"`
	// Job is the job_id of the current job.
	Job string `json:"job"`
	// Path is the path on the runner to the file that sets system PATH variables from workflow commands.
	Path string `json:"path"`
	// Ref is the fully-formed ref of the branch or tag that triggered the workflow run.
	Ref string `json:"ref"`
	// RefName is the short ref name of the branch or tag that triggered the workflow run.
	RefName string `json:"ref_name"`
	// RefProtected is true if branch protections or rulesets are configured for the ref.
	RefProtected bool `json:"ref_protected"`
	// RefType is the type of ref that triggered the workflow run (branch or tag).
	RefType string `json:"ref_type"`
	// Repository is the owner and repository name (e.g., octocat/Hello-World).
	Repository string `json:"repository"`
	// RepositoryID is the ID of the repository.
	RepositoryID string `json:"repository_id"`
	// RepositoryOwner is the repository owner's username.
	RepositoryOwner string `json:"repository_owner"`
	// RepositoryOwnerID is the repository owner's account ID.
	RepositoryOwnerID string `json:"repository_owner_id"`
	// RepositoryURL is the Git URL to the repository.
	RepositoryURL string `json:"repositoryUrl"`
	// RetentionDays is the number of days that workflow run logs and artifacts are kept.
	RetentionDays string `json:"retention_days"`
	// RunID is a unique number for each workflow run within a repository.
	RunID string `json:"run_id"`
	// RunNumber is a unique number for each run of a particular workflow in a repository.
	RunNumber string `json:"run_number"`
	// RunAttempt is a unique number for each attempt of a particular workflow run in a repository.
	RunAttempt string `json:"run_attempt"`
	// SecretSource is the source of a secret used in a workflow.
	SecretSource string `json:"secret_source"`
	// ServerURL is the URL of the GitHub server.
	ServerURL string `json:"server_url"`
	// Sha is the commit SHA that triggered the workflow.
	Sha string `json:"sha"`
	// Token is a token to authenticate on behalf of the GitHub App installed on your repository.
	Token string `json:"token"`
	// TriggeringActor is the username of the user that initiated the workflow run.
	TriggeringActor string `json:"triggering_actor"`
	// Workflow is the name of the workflow.
	Workflow string `json:"workflow"`
	// WorkflowRef is the ref path to the workflow.
	WorkflowRef string `json:"workflow_ref"`
	// WorkflowSha is the commit SHA for the workflow file.
	WorkflowSha string `json:"workflow_sha"`
	// Workspace is the default working directory on the runner for steps.
	Workspace string `json:"workspace"`
}

// NeedsContext represents the output and result of a job that the current job depends on.
type NeedsContext struct {
	// Outputs contains the set of outputs of the job.
	Outputs map[string]string `json:"outputs"`
	// Result is the result of the job (success, failure, cancelled, or skipped).
	Result string `json:"result"`
}

// JobContext contains information about the currently running job.
type JobContext struct {
	// CheckRunID is the check run ID of the current job.
	CheckRunID int64 `json:"check_run_id"`
	// Container contains information about the job's container.
	Container JobContextContainer `json:"container"`
	// Services contains information about the service containers created for a job.
	Services map[string]JobContextService `json:"services"`
	// Status is the current status of the job (success, failure, or cancelled).
	Status string `json:"status"`
}

// JobContextContainer contains information about a container in a job.
type JobContextContainer struct {
	// ID is the ID of the container.
	ID string `json:"id"`
	// Network is the ID of the container network.
	Network string `json:"network"`
}

// JobContextService contains information about a service container in a job.
type JobContextService struct {
	// ID is the ID of the service container.
	ID string `json:"id"`
	// Network is the ID of the service container network.
	Network string `json:"network"`
	// Ports contains the exposed ports of the service container.
	Ports map[string]string `json:"ports"`
}

// JobsContext contains information about jobs in a reusable workflow.
type JobsContext map[string]JobsContextEntry

// JobsContextEntry represents the result and outputs of a job in a reusable workflow.
type JobsContextEntry struct {
	// Result is the result of the job (success, failure, cancelled, or skipped).
	Result string `json:"result"`
	// Outputs contains the set of outputs of the job.
	Outputs map[string]string `json:"outputs"`
}

// StepsContext contains information about the steps in the current job that have an id specified and have already run.
type StepsContext map[string]StepsContextEntry

// StepsContextEntry represents the results and outputs of a step.
type StepsContextEntry struct {
	// Outputs contains the set of outputs defined for the step.
	Outputs map[string]string `json:"outputs"`
	// Conclusion is the result of a completed step after continue-on-error is applied.
	Conclusion string `json:"conclusion"`
	// Outcome is the result of a completed step before continue-on-error is applied.
	Outcome string `json:"outcome"`
}

// RunnerContext contains information about the runner that is executing the current job.
type RunnerContext struct {
	// Name is the name of the runner executing the job.
	Name string `json:"name"`
	// OS is the operating system of the runner executing the job.
	OS string `json:"os"`
	// Arch is the architecture of the runner executing the job.
	Arch string `json:"arch"`
	// Temp is the path to a temporary directory on the runner.
	Temp string `json:"temp"`
	// ToolCache is the path to the directory containing preinstalled tools for GitHub-hosted runners.
	ToolCache string `json:"tool_cache"`
	// Debug is set only if debug logging is enabled, and always has the value of 1.
	Debug string `json:"debug"`
	// Environment is the environment of the runner executing the job (github-hosted or self-hosted).
	Environment string `json:"environment"`
}

// SecretsContext contains the names and values of secrets that are available to a workflow run.
type SecretsContext map[string]string

// StrategyContext contains information about the matrix execution strategy for the current job.
type StrategyContext struct {
	// FailFast evaluates to true if all in-progress jobs are canceled if any job in a matrix fails.
	FailFast bool `json:"fail-fast"`
	// JobIndex is the index of the current job in the matrix (zero-based).
	JobIndex int `json:"job-index"`
	// JobTotal is the total number of jobs in the matrix.
	JobTotal int `json:"job-total"`
	// MaxParallel is the maximum number of jobs that can run simultaneously when using a matrix job strategy.
	MaxParallel int `json:"max-parallel"`
}
