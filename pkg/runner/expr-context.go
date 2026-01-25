package runner

import (
	"maps"

	"github.com/drornir/better-actions/pkg/runner/expr"
)

type MakeExprContextParams struct {
	GlobalEnv map[string]string
	Workflow  *WorkflowState
	Job       *Job
	Step      *StepContext
}

func MakeExprContext(p MakeExprContextParams) (*expr.EvalContext, error) {
	var env map[string]string
	if p.Step != nil {
		env = p.Step.Env
	} else if p.Job != nil {
		env = p.Job.StepsEnvCopy()
	} else if p.Workflow != nil {
		env = maps.Clone(p.Workflow.Env)
	} else if p.GlobalEnv != nil {
		env = maps.Clone(p.GlobalEnv)
	} else {
		env = map[string]string{}
	}

	return &expr.EvalContext{
		Github: expr.GithubContext{
			Action:            "",
			ActionPath:        "",
			ActionRef:         "",
			ActionRepository:  "",
			ActionStatus:      "",
			Actor:             "",
			ActorID:           "",
			APIURL:            "",
			BaseRef:           "",
			Env:               "",
			Event:             expr.JSObject{},
			EventName:         "",
			EventPath:         "",
			GraphQLURL:        "",
			HeadRef:           "",
			Job:               "",
			Path:              "",
			Ref:               "",
			RefName:           "",
			RefProtected:      false,
			RefType:           "",
			Repository:        "",
			RepositoryID:      "",
			RepositoryOwner:   "",
			RepositoryOwnerID: "",
			RepositoryURL:     "",
			RetentionDays:     "",
			RunID:             "",
			RunNumber:         "",
			RunAttempt:        "",
			SecretSource:      "",
			ServerURL:         "",
			Sha:               "",
			Token:             "",
			TriggeringActor:   "",
			Workflow:          "",
			WorkflowRef:       "",
			WorkflowSha:       "",
			Workspace:         "",
		},
		Env:      env,
		Job:      expr.JobContext{},
		Jobs:     expr.JobsContext{},
		Steps:    expr.StepsContext{},
		Runner:   expr.RunnerContext{},
		Secrets:  expr.SecretsContext{},
		Vars:     map[string]string{},
		Strategy: expr.StrategyContext{},
		Matrix:   expr.JSObject{},
		Needs:    map[string]expr.NeedsContext{},
		Inputs:   expr.JSObject{},
	}, nil
}
