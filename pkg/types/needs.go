package types

// Needs context contains the outputs of all jobs that are defined as a dependency of the current job.
type Needs struct {
	Jobs map[string]*NeedJobResult `json:"jobs"`
}

// NeedJobResult represents information about a job dependency.
type NeedJobResult struct {
	Result  string            `json:"result"`
	Outputs map[string]string `json:"outputs"`
}
