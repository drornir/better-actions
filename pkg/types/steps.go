package types

// Steps context contains information about the steps that have been run in the current job.
type Steps struct {
	Steps map[string]*StepResult `json:"steps"`
}

// StepResult represents information about a step.
type StepResult struct {
	Outputs    map[string]string `json:"outputs"`
	Conclusion string            `json:"conclusion"`
	Outcome    string            `json:"outcome"`
}
