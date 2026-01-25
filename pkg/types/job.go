package types

// Job context contains information about the currently running job.
type Job struct {
	CheckRunID int64               `json:"check_run_id"`
	Container  *Container          `json:"container,omitempty"`
	Services   map[string]*Service `json:"services,omitempty"`
	Status     string              `json:"status"`
}

// Container represents job container information.
type Container struct {
	ID      string `json:"id"`
	Network string `json:"network"`
}

// Service represents a service container in a job.
type Service struct {
	ID      string            `json:"id"`
	Network string            `json:"network"`
	Ports   map[string]string `json:"ports"`
}
