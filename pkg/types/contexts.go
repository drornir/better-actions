package types

// Env context contains variables set in a workflow, job, or step.
type Env map[string]string

// Vars context contains custom configuration variables set at the organization, repository, and environment levels.
type Vars map[string]string

// Secrets context contains the names and values of secrets available to a workflow run.
type Secrets map[string]string

// Inputs context contains the inputs of a reusable or manually triggered workflow.
type Inputs map[string]any

// AllContexts represents all available GitHub Actions contexts.
type AllContexts struct {
	GitHub   *GitHub   `json:"github,omitempty"`
	Env      Env       `json:"env,omitempty"`
	Vars     Vars      `json:"vars,omitempty"`
	Job      *Job      `json:"job,omitempty"`
	Jobs     *Jobs     `json:"jobs,omitempty"`
	Steps    *Steps    `json:"steps,omitempty"`
	Runner   *Runner   `json:"runner,omitempty"`
	Secrets  Secrets   `json:"secrets,omitempty"`
	Strategy *Strategy `json:"strategy,omitempty"`
	Matrix   Matrix    `json:"matrix,omitempty"`
	Needs    *Needs    `json:"needs,omitempty"`
	Inputs   Inputs    `json:"inputs,omitempty"`
}

type WorkflowContexts struct {
	GitHub  *GitHub `json:"github,omitempty"`
	Env     Env     `json:"env,omitempty"`
	Vars    Vars    `json:"vars,omitempty"`
	Runner  *Runner `json:"runner,omitempty"`
	Secrets Secrets `json:"secrets,omitempty"`
	Inputs  Inputs  `json:"inputs,omitempty"`
}
