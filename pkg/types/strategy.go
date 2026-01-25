package types

// Strategy context contains information about the matrix execution strategy for the current job.
type Strategy struct {
	FailFast    bool `json:"fail-fast"`
	JobIndex    int  `json:"job-index"`
	JobTotal    int  `json:"job-total"`
	MaxParallel int  `json:"max-parallel"`
}

// Matrix context contains the matrix properties defined in the workflow that apply to the current job.
type Matrix map[string]any
