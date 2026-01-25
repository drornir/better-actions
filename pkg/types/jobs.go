package types

// Jobs context is only available in reusable workflows.
type Jobs struct {
	Jobs map[string]*JobResult `json:"jobs"`
}

// JobResult represents the result of a job in a reusable workflow.
type JobResult struct {
	Result  string            `json:"result"`
	Outputs map[string]string `json:"outputs"`
}
