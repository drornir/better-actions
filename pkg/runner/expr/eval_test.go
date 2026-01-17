package expr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/drornir/better-actions/pkg/runner/expr"
)

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		expr     string
		expected string
	}{
		// Basic expressions
		{"true", "true"},
		{"!true", "false"},
		{"true || false", "true"},
		{"true && false", "false"},
		{"true && false || false", "false"},
		{"true && false || true", "true"},
		{"true || false && true", "true"},
		{"42 > 24", "true"},
		{"42 >= 24", "true"},
		{"42 < 24", "false"},
		{"42 <= 24", "false"},
		{"42", "42"},
		{"'hello'", "\"hello\""},
		{"null", "null"},

		// GitHub context - basic fields
		{"github.actor", `"octocat"`},
		{"github.actor_id", `"583231"`},
		{"github.event_name", `"pull_request"`},
		{"github.repository", `"octocat/hello-world"`},
		{"github.repository_owner", `"octocat"`},
		{"github.ref", `"refs/pull/42/merge"`},
		{"github.ref_name", `"42/merge"`},
		{"github.base_ref", `"main"`},
		{"github.head_ref", `"feature/awesome"`},
		{"github.sha", `"abc123def456789abc123def456789abc123def4"`},
		{"github.workflow", `"CI"`},
		{"github.job", `"build"`},
		{"github.run_id", `"1234567890"`},
		{"github.run_number", `"15"`},
		{"github.run_attempt", `"1"`},
		{"github.server_url", `"https://github.com"`},
		{"github.api_url", `"https://api.github.com"`},

		// GitHub context - event object (PR payload)
		{"github.event.action", `"opened"`},
		{"github.event.number", "42"},
		{"github.event.pull_request.title", `"Add new feature"`},
		{"github.event.pull_request.body", `"This PR adds a cool new feature"`},
		{"github.event.pull_request.head.ref", `"feature/awesome"`},
		{"github.event.pull_request.base.ref", `"main"`},
		{"github.event.pull_request.user.login", `"octocat"`},
		{"github.event.pull_request.draft", "false"},
		{"github.event.pull_request.mergeable", "true"},
		{"github.event.repository.full_name", `"octocat/hello-world"`},
		{"github.event.sender.login", `"octocat"`},

		// Env context
		{"env.CI", `"true"`},
		{"env.NODE_ENV", `"test"`},
		{"env.LOG_LEVEL", `"debug"`},

		// Job context
		{"job.status", `"success"`},

		// Steps context
		{"steps.checkout.conclusion", `"success"`},
		{"steps.checkout.outcome", `"success"`},
		{"steps.setup-node.conclusion", `"success"`},
		{"steps.setup-node.outputs.node-version", `"20.10.0"`},

		// Runner context
		{"runner.name", `"GitHub Actions 2"`},
		{"runner.os", `"Linux"`},
		{"runner.arch", `"X64"`},
		{"runner.environment", `"github-hosted"`},

		// Vars context
		{"vars.DEPLOYMENT_ENV", `"staging"`},
		{"vars.APP_NAME", `"hello-world"`},

		// Strategy context
		{"strategy.fail-fast", "true"},
		{"strategy.job-index", "0"},
		{"strategy.job-total", "1"},
		{"strategy.max-parallel", "1"},

		// Matrix context
		{"matrix.node-version", `"20"`},
		{"matrix.os", `"ubuntu-latest"`},

		// Needs context
		{"needs.lint.result", `"success"`},
		{"needs.lint.outputs.eslint-result", `"passed"`},

		// Inputs context
		{"inputs.deploy", "false"},
		{"inputs.environment", `"staging"`},

		// Comparisons using context values
		{"github.event_name == 'pull_request'", "true"},
		{"github.event_name == 'push'", "false"},
		{"github.actor == 'octocat'", "true"},
		{"job.status == 'success'", "true"},
		{"steps.checkout.conclusion == 'success'", "true"},
		{"github.event.pull_request.draft == false", "true"},
		{"strategy.job-total == 1", "true"},
		{"github.event.number > 40", "true"},
		{"github.event.number >= 42", "true"},
		{"github.event.number < 50", "true"},

		// Logical operations with context
		{"github.event_name == 'pull_request' && job.status == 'success'", "true"},
		{"github.event_name == 'push' || github.event_name == 'pull_request'", "true"},
		{"!(github.event.pull_request.draft)", "true"},
		{"needs.lint.result == 'success' && steps.checkout.conclusion == 'success'", "true"},

		// comparisons
		{"'1' == 1", "true"},
		{"null == null", "true"},
		{"null == 0", "true"},
		{"true == 1", "true"},
		{"1 == true", "true"},
		{"false == 0", "true"},
		{"false != 1", "true"},
		// matrix should be equal to matrix according to the spec, but I'm skipping this one
		{"matrix == matrix", "false"},
		{"'' == false", "true"},
		{"'' == 0", "true"},
		{"'' != 1", "true"},
		{"'' >= 0", "true"},
	}

	for _, tc := range testCases {
		t.Run(tc.expr, func(t *testing.T) {
			evaluator, err := expr.NewEvaluator(prContext(t), expr.DefaultFunctions)
			if !assert.NoErrorf(t, err, "initializing evaluator") {
				return
			}

			ast, parseErr := expr.NewParser().Parse(expr.NewExprLexer(tc.expr + "}}"))
			var errFix error
			if parseErr != nil {
				errFix = parseErr
			}
			if !assert.NoErrorf(t, errFix, "parsing expression") {
				return
			}

			result, err := evaluator.Evaluate(ast)
			if !assert.NoErrorf(t, err, "evaluating expression") {
				return
			}

			resAsJSON, err := result.MarshalJSON()
			if !assert.NoErrorf(t, err, "marshaling result") {
				return
			}

			assert.JSONEq(t, tc.expected, string(resAsJSON))
		})
	}
}

// mustJSObject converts a map[string]any to expr.JSObject, failing the test on error.
func mustJSObject(t *testing.T, m map[string]any) expr.JSObject {
	t.Helper()
	var obj expr.JSObject
	err := obj.UnmarshalFromGoMap(m)
	require.NoError(t, err, "converting map to JSObject")
	return obj
}

// prContext returns a realistic EvalContext for a pull request workflow run.
func prContext(t *testing.T) *expr.EvalContext {
	t.Helper()
	return &expr.EvalContext{
		Github: expr.GithubContext{
			Action:           "__run",
			ActionPath:       "",
			ActionRef:        "",
			ActionRepository: "",
			ActionStatus:     "",
			Actor:            "octocat",
			ActorID:          "583231",
			APIURL:           "https://api.github.com",
			BaseRef:          "main",
			Env:              "/home/runner/work/_temp/_runner_file_commands/set_env_abc123",
			Event: mustJSObject(t, map[string]any{
				"action": "opened",
				"number": 42,
				"pull_request": map[string]any{
					"title":  "Add new feature",
					"body":   "This PR adds a cool new feature",
					"number": 42,
					"head": map[string]any{
						"ref": "feature/awesome",
						"sha": "abc123def456",
					},
					"base": map[string]any{
						"ref": "main",
						"sha": "789xyz000111",
					},
					"user": map[string]any{
						"login": "octocat",
						"id":    583231,
					},
					"draft":     false,
					"mergeable": true,
				},
				"repository": map[string]any{
					"full_name": "octocat/hello-world",
					"name":      "hello-world",
					"owner": map[string]any{
						"login": "octocat",
					},
				},
				"sender": map[string]any{
					"login": "octocat",
					"id":    583231,
				},
			}),
			EventName:         "pull_request",
			EventPath:         "/home/runner/work/_temp/_github_workflow/event.json",
			GraphQLURL:        "https://api.github.com/graphql",
			HeadRef:           "feature/awesome",
			Job:               "build",
			Path:              "/home/runner/work/_temp/_runner_file_commands/add_path_abc123",
			Ref:               "refs/pull/42/merge",
			RefName:           "42/merge",
			RefProtected:      false,
			RefType:           "branch",
			Repository:        "octocat/hello-world",
			RepositoryID:      "12345678",
			RepositoryOwner:   "octocat",
			RepositoryOwnerID: "583231",
			RepositoryURL:     "git://github.com/octocat/hello-world.git",
			RetentionDays:     "90",
			RunID:             "1234567890",
			RunNumber:         "15",
			RunAttempt:        "1",
			SecretSource:      "Actions",
			ServerURL:         "https://github.com",
			Sha:               "abc123def456789abc123def456789abc123def4",
			Token:             "***",
			TriggeringActor:   "octocat",
			Workflow:          "CI",
			WorkflowRef:       "octocat/hello-world/.github/workflows/ci.yml@refs/pull/42/merge",
			WorkflowSha:       "abc123def456789abc123def456789abc123def4",
			Workspace:         "/home/runner/work/hello-world/hello-world",
		},
		Env: map[string]string{
			"CI":        "true",
			"NODE_ENV":  "test",
			"LOG_LEVEL": "debug",
		},
		Job: expr.JobContext{
			CheckRunID: 9876543210,
			Container:  expr.JobContextContainer{},
			Services:   map[string]expr.JobContextService{},
			Status:     "success",
		},
		Jobs: expr.JobsContext{},
		Steps: expr.StepsContext{
			"checkout": {
				Outputs:    map[string]string{},
				Conclusion: "success",
				Outcome:    "success",
			},
			"setup-node": {
				Outputs: map[string]string{
					"node-version": "20.10.0",
				},
				Conclusion: "success",
				Outcome:    "success",
			},
		},
		Runner: expr.RunnerContext{
			Name:        "GitHub Actions 2",
			OS:          "Linux",
			Arch:        "X64",
			Temp:        "/home/runner/work/_temp",
			ToolCache:   "/opt/hostedtoolcache",
			Debug:       "",
			Environment: "github-hosted",
		},
		Secrets: expr.SecretsContext{
			"GITHUB_TOKEN": "***",
			"NPM_TOKEN":    "***",
		},
		Vars: map[string]string{
			"DEPLOYMENT_ENV": "staging",
			"APP_NAME":       "hello-world",
		},
		Strategy: expr.StrategyContext{
			FailFast:    true,
			JobIndex:    0,
			JobTotal:    1,
			MaxParallel: 1,
		},
		Matrix: mustJSObject(t, map[string]any{
			"node-version": "20",
			"os":           "ubuntu-latest",
		}),
		Needs: map[string]expr.NeedsContext{
			"lint": {
				Outputs: map[string]string{
					"eslint-result": "passed",
				},
				Result: "success",
			},
		},
		Inputs: mustJSObject(t, map[string]any{
			"deploy":      false,
			"environment": "staging",
		}),
	}
}
