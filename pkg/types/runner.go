package types

// Runner context contains information about the runner executing the current job.
type Runner struct {
	Name        string `json:"name"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Temp        string `json:"temp"`
	ToolCache   string `json:"tool_cache"`
	Debug       string `json:"debug"`
	Environment string `json:"environment"`
}
