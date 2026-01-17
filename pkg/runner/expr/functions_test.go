package expr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/drornir/better-actions/pkg/runner/expr"
)

func TestFuncContains(t *testing.T) {
	testCases := []struct {
		name     string
		args     []expr.JSValue
		expected bool
		wantErr  bool
	}{
		{
			name: "string contains substring",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("llo")},
			},
			expected: true,
		},
		{
			name: "string contains substring case-insensitive",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("LLO")},
			},
			expected: true,
		},
		{
			name: "string does not contain substring",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("xyz")},
			},
			expected: false,
		},
		{
			name: "array contains string element",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("push")},
					{String: expr.Some("pull_request")},
				})},
				{String: expr.Some("push")},
			},
			expected: true,
		},
		{
			name: "array contains string element case-insensitive",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("bug")},
					{String: expr.Some("help wanted")},
				})},
				{String: expr.Some("BUG")},
			},
			expected: true,
		},
		{
			name: "array does not contain element",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("push")},
					{String: expr.Some("pull_request")},
				})},
				{String: expr.Some("workflow_dispatch")},
			},
			expected: false,
		},
		{
			name: "empty string contains empty string",
			args: []expr.JSValue{
				{String: expr.Some("")},
				{String: expr.Some("")},
			},
			expected: true,
		},
		{
			name: "null value casts to empty string",
			args: []expr.JSValue{
				{String: expr.Some("hello")},
				{Null: expr.Some(struct{}{})},
			},
			expected: true, // empty string is contained in any string
		},
		{
			name: "too few arguments",
			args: []expr.JSValue{
				{String: expr.Some("hello")},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("contains")
			require.True(t, ok, "contains function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.Boolean.IsPresent, "result should be a boolean")
			assert.Equal(t, tc.expected, result.Boolean.Value)
		})
	}
}

func TestFuncStartsWith(t *testing.T) {
	testCases := []struct {
		name     string
		args     []expr.JSValue
		expected bool
		wantErr  bool
	}{
		{
			name: "string starts with prefix",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("He")},
			},
			expected: true,
		},
		{
			name: "string starts with prefix case-insensitive",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("HELLO")},
			},
			expected: true,
		},
		{
			name: "string does not start with prefix",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("world")},
			},
			expected: false,
		},
		{
			name: "empty prefix matches any string",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("")},
			},
			expected: true,
		},
		{
			name: "number cast to string",
			args: []expr.JSValue{
				{Float: expr.Some(float64(12345))},
				{String: expr.Some("123")},
			},
			expected: true,
		},
		{
			name: "too few arguments",
			args: []expr.JSValue{
				{String: expr.Some("hello")},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("startsWith")
			require.True(t, ok, "startsWith function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.Boolean.IsPresent, "result should be a boolean")
			assert.Equal(t, tc.expected, result.Boolean.Value)
		})
	}
}

func TestFuncEndsWith(t *testing.T) {
	testCases := []struct {
		name     string
		args     []expr.JSValue
		expected bool
		wantErr  bool
	}{
		{
			name: "string ends with suffix",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("ld")},
			},
			expected: true,
		},
		{
			name: "string ends with suffix case-insensitive",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("WORLD")},
			},
			expected: true,
		},
		{
			name: "string does not end with suffix",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("Hello")},
			},
			expected: false,
		},
		{
			name: "empty suffix matches any string",
			args: []expr.JSValue{
				{String: expr.Some("Hello world")},
				{String: expr.Some("")},
			},
			expected: true,
		},
		{
			name: "boolean cast to string",
			args: []expr.JSValue{
				{String: expr.Some("result: true")},
				{Boolean: expr.Some(true)},
			},
			expected: true,
		},
		{
			name: "too few arguments",
			args: []expr.JSValue{
				{String: expr.Some("hello")},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("endsWith")
			require.True(t, ok, "endsWith function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.Boolean.IsPresent, "result should be a boolean")
			assert.Equal(t, tc.expected, result.Boolean.Value)
		})
	}
}

func TestFuncFormat(t *testing.T) {
	testCases := []struct {
		name     string
		args     []expr.JSValue
		expected string
		wantErr  bool
	}{
		{
			name: "simple format with three placeholders",
			args: []expr.JSValue{
				{String: expr.Some("Hello {0} {1} {2}")},
				{String: expr.Some("Mona")},
				{String: expr.Some("the")},
				{String: expr.Some("Octocat")},
			},
			expected: "Hello Mona the Octocat",
		},
		{
			name: "format with escaped braces",
			args: []expr.JSValue{
				{String: expr.Some("{{Hello {0} {1} {2}!}}")},
				{String: expr.Some("Mona")},
				{String: expr.Some("the")},
				{String: expr.Some("Octocat")},
			},
			expected: "{Hello Mona the Octocat!}",
		},
		{
			name: "format with repeated placeholder",
			args: []expr.JSValue{
				{String: expr.Some("{0} and {0}")},
				{String: expr.Some("test")},
			},
			expected: "test and test",
		},
		{
			name: "format with number value",
			args: []expr.JSValue{
				{String: expr.Some("count: {0}")},
				{Float: expr.Some(float64(42))},
			},
			expected: "count: 42",
		},
		{
			name: "format with boolean value",
			args: []expr.JSValue{
				{String: expr.Some("enabled: {0}")},
				{Boolean: expr.Some(true)},
			},
			expected: "enabled: true",
		},
		{
			name: "format without placeholders",
			args: []expr.JSValue{
				{String: expr.Some("no placeholders")},
				{String: expr.Some("unused")},
			},
			expected: "no placeholders",
		},
		{
			name: "too few arguments",
			args: []expr.JSValue{
				{String: expr.Some("hello {0}")},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("format")
			require.True(t, ok, "format function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.String.IsPresent, "result should be a string")
			assert.Equal(t, tc.expected, result.String.Value)
		})
	}
}

func TestFuncJoin(t *testing.T) {
	testCases := []struct {
		name     string
		args     []expr.JSValue
		expected string
		wantErr  bool
	}{
		{
			name: "join array with default separator",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("a")},
					{String: expr.Some("b")},
					{String: expr.Some("c")},
				})},
			},
			expected: "a,b,c",
		},
		{
			name: "join array with custom separator",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("bug")},
					{String: expr.Some("help wanted")},
				})},
				{String: expr.Some(", ")},
			},
			expected: "bug, help wanted",
		},
		{
			name: "join empty array",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{})},
			},
			expected: "",
		},
		{
			name: "join single element array",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("only")},
				})},
			},
			expected: "only",
		},
		{
			name: "join array with mixed types",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("text")},
					{Float: expr.Some(float64(42))},
					{Boolean: expr.Some(true)},
				})},
				{String: expr.Some("-")},
			},
			expected: "text-42-true",
		},
		{
			name: "join string returns itself",
			args: []expr.JSValue{
				{String: expr.Some("hello")},
			},
			expected: "hello",
		},
		{
			name:    "no arguments",
			args:    []expr.JSValue{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("join")
			require.True(t, ok, "join function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.String.IsPresent, "result should be a string")
			assert.Equal(t, tc.expected, result.String.Value)
		})
	}
}

func TestFuncToJSON(t *testing.T) {
	testCases := []struct {
		name         string
		args         []expr.JSValue
		expectedJSON string
		wantErr      bool
	}{
		{
			name: "object to JSON",
			args: []expr.JSValue{
				{Object: expr.Some(expr.JSObject{
					"status": {String: expr.Some("success")},
					"aNull":  {Null: expr.Some(struct{}{})},
				})},
			},
			expectedJSON: `{"status":"success", "aNull":null}`,
		},
		{
			name: "string to JSON",
			args: []expr.JSValue{
				{String: expr.Some("hello")},
			},
			expectedJSON: `"hello"`,
		},
		{
			name: "number to JSON",
			args: []expr.JSValue{
				{Float: expr.Some(float64(42))},
			},
			expectedJSON: "42",
		},
		{
			name: "boolean to JSON",
			args: []expr.JSValue{
				{Boolean: expr.Some(true)},
			},
			expectedJSON: "true",
		},
		{
			name: "null to JSON",
			args: []expr.JSValue{
				{Null: expr.Some(struct{}{})},
			},
			expectedJSON: "null",
		},
		{
			name: "array to JSON",
			args: []expr.JSValue{
				{Array: expr.Some(expr.JSArray{
					{String: expr.Some("a")},
					{String: expr.Some("b")},
				})},
			},
			expectedJSON: `["a","b"]`,
		},
		{
			name:    "no arguments",
			args:    []expr.JSValue{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("toJSON")
			require.True(t, ok, "toJSON function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.String.IsPresent, "result should be a string")

			assert.JSONEq(t, tc.expectedJSON, result.String.Value)
		})
	}
}

func TestFuncFromJSON(t *testing.T) {
	testCases := []struct {
		name         string
		args         []expr.JSValue
		expectedType expr.JSValueType
		checkValue   func(t *testing.T, v expr.JSValue)
		wantErr      bool
	}{
		{
			name: "parse object",
			args: []expr.JSValue{
				{String: expr.Some(`{"status": "success"}`)},
			},
			expectedType: expr.ObjectType,
			checkValue: func(t *testing.T, v expr.JSValue) {
				status, ok := v.Object.Value["status"]
				require.True(t, ok, "status key should exist")
				assert.Equal(t, "success", status.String.Value)
			},
		},
		{
			name: "parse array",
			args: []expr.JSValue{
				{String: expr.Some(`["push", "pull_request"]`)},
			},
			expectedType: expr.ArrayType,
			checkValue: func(t *testing.T, v expr.JSValue) {
				require.Len(t, v.Array.Value, 2)
				assert.Equal(t, "push", v.Array.Value[0].String.Value)
				assert.Equal(t, "pull_request", v.Array.Value[1].String.Value)
			},
		},
		{
			name: "parse string",
			args: []expr.JSValue{
				{String: expr.Some(`"hello"`)},
			},
			expectedType: expr.StringType,
			checkValue: func(t *testing.T, v expr.JSValue) {
				assert.Equal(t, "hello", v.String.Value)
			},
		},
		{
			name: "parse number",
			args: []expr.JSValue{
				{String: expr.Some("42")},
			},
			expectedType: expr.NumberType,
			checkValue: func(t *testing.T, v expr.JSValue) {
				assert.Equal(t, float64(42), v.Float.Value)
			},
		},
		{
			name: "parse boolean true",
			args: []expr.JSValue{
				{String: expr.Some("true")},
			},
			expectedType: expr.BooleanType,
			checkValue: func(t *testing.T, v expr.JSValue) {
				assert.True(t, v.Boolean.Value)
			},
		},
		{
			name: "parse boolean false",
			args: []expr.JSValue{
				{String: expr.Some("false")},
			},
			expectedType: expr.BooleanType,
			checkValue: func(t *testing.T, v expr.JSValue) {
				assert.False(t, v.Boolean.Value)
			},
		},
		{
			name: "parse null",
			args: []expr.JSValue{
				{String: expr.Some("null")},
			},
			expectedType: expr.NullType,
		},
		{
			name: "parse complex matrix",
			args: []expr.JSValue{
				{String: expr.Some(`{"include":[{"project":"foo","config":"Debug"},{"project":"bar","config":"Release"}]}`)},
			},
			expectedType: expr.ObjectType,
			checkValue: func(t *testing.T, v expr.JSValue) {
				include, ok := v.Object.Value["include"]
				require.True(t, ok)
				require.True(t, include.Array.IsPresent)
				require.Len(t, include.Array.Value, 2)
				require.True(t, include.Array.Value[0].Object.IsPresent)
				require.True(t, include.Array.Value[1].Object.IsPresent)
				assert.Equal(t, "foo", include.Array.Value[0].Object.Value["project"].String.Value)
				assert.Equal(t, "Release", include.Array.Value[1].Object.Value["config"].String.Value)
			},
		},
		{
			name: "invalid JSON",
			args: []expr.JSValue{
				{String: expr.Some("{invalid}")},
			},
			wantErr: true,
		},
		{
			name:    "no arguments",
			args:    []expr.JSValue{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("fromJSON")
			require.True(t, ok, "fromJSON function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedType, result.Type())

			if tc.checkValue != nil {
				tc.checkValue(t, result)
			}
		})
	}
}

func TestFuncHashFiles(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"file1.txt":           "content of file 1",
		"file2.txt":           "content of file 2",
		"src/main.go":         "package main",
		"src/util.go":         "package util",
		"src/sub/nested.go":   "package nested",
		"lib/foo.rb":          "ruby code",
		"lib/bar.rb":          "more ruby",
		"lib/foo/excluded.rb": "should be excluded",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0o644)
		require.NoError(t, err)
	}

	// Set the workspace for testing
	originalWorkspace := os.Getenv("GITHUB_WORKSPACE")
	t.Cleanup(func() {
		os.Setenv("GITHUB_WORKSPACE", originalWorkspace)
	})
	os.Setenv("GITHUB_WORKSPACE", tmpDir)

	testCases := []struct {
		name              string
		args              []expr.JSValue
		expectEmpty       bool
		expectConsistent  bool
		compareWithSecond []expr.JSValue
		wantErr           bool
	}{
		{
			name: "hash single file",
			args: []expr.JSValue{
				{String: expr.Some("file1.txt")},
			},
			expectConsistent: true,
		},
		{
			name: "hash multiple files with wildcard",
			args: []expr.JSValue{
				{String: expr.Some("*.txt")},
			},
			expectConsistent: true,
		},
		{
			name: "hash with double star pattern",
			args: []expr.JSValue{
				{String: expr.Some("**/*.go")},
			},
			expectConsistent: true,
		},
		{
			name: "no matching files returns empty string",
			args: []expr.JSValue{
				{String: expr.Some("nonexistent/*.xyz")},
			},
			expectEmpty: true,
		},
		{
			name: "different files produce different hashes",
			args: []expr.JSValue{
				{String: expr.Some("file1.txt")},
			},
			compareWithSecond: []expr.JSValue{
				{String: expr.Some("file2.txt")},
			},
		},
		{
			name: "multiple patterns",
			args: []expr.JSValue{
				{String: expr.Some("file1.txt")},
				{String: expr.Some("file2.txt")},
			},
			expectConsistent: true,
		},
		{
			name:    "no arguments",
			args:    []expr.JSValue{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get("hashFiles")
			require.True(t, ok, "hashFiles function should exist")

			result, err := fn(tc.args...)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.String.IsPresent, "result should be a string")

			if tc.expectEmpty {
				assert.Empty(t, result.String.Value)
				return
			}

			// Hash should be a valid hex string (SHA-256 = 64 hex chars)
			if result.String.Value != "" {
				assert.Len(t, result.String.Value, 64, "SHA-256 hash should be 64 hex chars")
			}

			if tc.expectConsistent {
				// Run again to verify consistency
				result2, err := fn(tc.args...)
				require.NoError(t, err)
				assert.Equal(t, result.String.Value, result2.String.Value, "hash should be consistent")
			}

			if tc.compareWithSecond != nil {
				result2, err := fn(tc.compareWithSecond...)
				require.NoError(t, err)
				assert.NotEqual(t, result.String.Value, result2.String.Value, "different files should produce different hashes")
			}
		})
	}
}

func TestFunctionStoreCaseInsensitive(t *testing.T) {
	testCases := []struct {
		name     string
		funcName string
	}{
		{"lowercase contains", "contains"},
		{"uppercase CONTAINS", "CONTAINS"},
		{"mixed case Contains", "Contains"},
		{"lowercase startswith", "startswith"},
		{"camelCase startsWith", "startsWith"},
		{"lowercase fromjson", "fromjson"},
		{"camelCase fromJSON", "fromJSON"},
		{"uppercase TOJSON", "TOJSON"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get(tc.funcName)
			assert.True(t, ok, "function %s should be found", tc.funcName)
			assert.NotNil(t, fn)
		})
	}
}

func TestUnimplementedFunctions(t *testing.T) {
	unimplementedFuncs := []string{
		"success",
		"always",
		"cancelled",
		"failure",
	}

	for _, name := range unimplementedFuncs {
		t.Run(name, func(t *testing.T) {
			fn, ok := expr.DefaultFunctions.Get(name)
			require.True(t, ok, "%s function should exist", name)

			_, err := fn()
			assert.Error(t, err, "%s should return an error", name)
			assert.Contains(t, err.Error(), "not implemented")
		})
	}
}
