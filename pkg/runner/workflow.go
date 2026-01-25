package runner

import (
	"context"
	"maps"

	"github.com/drornir/better-actions/pkg/ctxkit"
	"github.com/drornir/better-actions/pkg/runner/expr"
	"github.com/drornir/better-actions/pkg/yamls"
)

type RunWorkflowParams struct {
	Github TODO
	Event  TODO
	Inputs TODO
}

func (r *Runner) RunWorkflow(ctx context.Context, wf *yamls.Workflow, params RunWorkflowParams) (*WorkflowState, error) {
	ctx, _, oopser := ctxkit.With(ctx, "workflow", wf.Name)
	jobs := wf.Jobs

	wfState := &WorkflowState{
		Name:   wf.Name,
		Jobs:   make(map[string]*Job, len(jobs)),
		Env:    nil, // need to run through tempalting
		Inputs: params.Inputs,
	}

	{
		exprContext, err := MakeExprContext(MakeExprContextParams{GlobalEnv: r.Env, Workflow: wfState})
		if err != nil {
			return nil, oopser.Wrapf(err, "failed to create expression context")
		}
		evaluator, err := expr.NewEvaluator(exprContext)
		if err != nil {
			return nil, oopser.Wrapf(err, "failed to create expression evaluator")
		}
		wfEnv := maps.Clone(r.Env)
		for k, v := range wf.Env {
			evaled, err := evaluator.EvaluateTemplate(v)
			if err != nil {
				return nil, oopser.Wrapf(err, "failed to evaluate env var %s", k)
			}
			wfEnv[k] = evaled
		}
		wfState.Env = wfEnv
	}

	for jobName, job := range jobs {
		j := NewJob(jobName, job, wfState, r.Console)
		wfState.Jobs[jobName] = j
	}

	// TODO parallel, remote execution etc.
	for _, j := range wfState.Jobs {
		err := j.Run(ctx)
		if err != nil {
			return nil, oopser.With("job", j.Name).Wrapf(err, "failed to run job")
		}
	}

	return wfState, nil
}

type WorkflowState struct {
	Name   string
	Jobs   map[string]*Job
	Env    map[string]string
	Inputs TODO
}
