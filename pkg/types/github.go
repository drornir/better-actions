package types

// GitHub context contains information about the workflow run and the event that triggered the run.
type GitHub struct {
	Action            string         `json:"action"`
	ActionPath        string         `json:"action_path"`
	ActionRef         string         `json:"action_ref"`
	ActionRepository  string         `json:"action_repository"`
	ActionStatus      string         `json:"action_status"`
	Actor             string         `json:"actor"`
	ActorID           string         `json:"actor_id"`
	APIURL            string         `json:"api_url"`
	BaseRef           string         `json:"base_ref"`
	Env               string         `json:"env"`
	Event             map[string]any `json:"event"`
	EventName         string         `json:"event_name"`
	EventPath         string         `json:"event_path"`
	GraphQLURL        string         `json:"graphql_url"`
	HeadRef           string         `json:"head_ref"`
	Job               string         `json:"job"`
	Path              string         `json:"path"`
	Ref               string         `json:"ref"`
	RefName           string         `json:"ref_name"`
	RefProtected      bool           `json:"ref_protected"`
	RefType           string         `json:"ref_type"`
	Repository        string         `json:"repository"`
	RepositoryID      string         `json:"repository_id"`
	RepositoryOwner   string         `json:"repository_owner"`
	RepositoryOwnerID string         `json:"repository_owner_id"`
	RepositoryURL     string         `json:"repositoryUrl"`
	RetentionDays     string         `json:"retention_days"`
	RunID             string         `json:"run_id"`
	RunNumber         string         `json:"run_number"`
	RunAttempt        string         `json:"run_attempt"`
	SecretSource      string         `json:"secret_source"`
	ServerURL         string         `json:"server_url"`
	SHA               string         `json:"sha"`
	Token             string         `json:"token"`
	TriggeringActor   string         `json:"triggering_actor"`
	Workflow          string         `json:"workflow"`
	WorkflowRef       string         `json:"workflow_ref"`
	WorkflowSHA       string         `json:"workflow_sha"`
	Workspace         string         `json:"workspace"`
}
