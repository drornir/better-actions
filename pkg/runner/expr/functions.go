package expr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/samber/oops"
)

type (
	// Function is a wrapper around the non-status (aka normal) functions
	// you can call from templates.
	Function      func(args ...JSValue) (JSValue, error)
	FunctionStore map[string]Function
)

var DefaultFunctions = FunctionStore{}

func init() {
	DefaultFunctions.Add("contains", funcContains)
	DefaultFunctions.Add("startsWith", funcStartsWith)
	DefaultFunctions.Add("endsWith", funcEndsWith)
	DefaultFunctions.Add("format", funcFormat)
	DefaultFunctions.Add("join", funcJoin)
	DefaultFunctions.Add("toJSON", funcToJSON)
	DefaultFunctions.Add("fromJSON", funcFromJSON)
	DefaultFunctions.Add("hashFiles", funcHashFiles)
}

func (fs FunctionStore) Add(name string, f Function) {
	fs[strings.ToLower(name)] = f
}

func (fs FunctionStore) Get(name string) (Function, bool) {
	f, ok := fs[strings.ToLower(name)]
	return f, ok
}

func funcTODOUnimplemented(_ ...JSValue) (JSValue, error) {
	return JSValue{}, oops.Errorf("function is not implemented")
}

// castToString casts a JSValue to a string according to GitHub Actions rules:
// - Null -> ”
// - Boolean -> 'true' or 'false'
// - Number -> Decimal format, exponential for large numbers
// - Array -> Arrays are not converted to a string (returns error)
// - Object -> Objects are not converted to a string (returns error)
func castToString(v JSValue) (string, error) {
	switch {
	case v.Null.IsPresent:
		return "", nil
	case v.Undefined.IsPresent:
		return "", nil
	case v.Boolean.IsPresent:
		if v.Boolean.Value {
			return "true", nil
		}
		return "false", nil
	case v.String.IsPresent:
		return v.String.Value, nil
	case v.Int.IsPresent:
		return strconv.FormatInt(v.Int.Value, 10), nil
	case v.Float.IsPresent:
		// Use exponential format for very large or very small numbers
		f := v.Float.Value
		if f != 0 && (f >= 1e15 || f <= -1e15 || (f < 1e-4 && f > -1e-4)) {
			return strconv.FormatFloat(f, 'e', -1, 64), nil
		}
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	case v.Array.IsPresent:
		return "", oops.Errorf("cannot convert array to string")
	case v.Object.IsPresent:
		return "", oops.Errorf("cannot convert object to string")
	default:
		return "", oops.Errorf("cannot convert unknown type to string")
	}
}

// funcContains implements the contains(search, item) function.
// Returns true if search contains item.
// If search is an array, returns true if item is an element in the array.
// If search is a string, returns true if item is a substring.
// This function is not case sensitive.
func funcContains(args ...JSValue) (JSValue, error) {
	if len(args) < 2 {
		return JSValue{}, oops.Errorf("contains requires 2 arguments, got %d", len(args))
	}

	search := args[0]
	item := args[1]

	// If search is an array, check if item is in the array
	if search.Array.IsPresent {
		itemStr, itemStrErr := castToString(item)

		for _, elem := range search.Array.Value {
			// Try direct comparison first for same types
			if equalJSValues(elem, item) {
				return JSValue{Boolean: Some(true)}, nil
			}

			// Fall back to string comparison (case-insensitive) if both can be cast
			elemStr, elemStrErr := castToString(elem)
			if itemStrErr == nil && elemStrErr == nil {
				if strings.EqualFold(elemStr, itemStr) {
					return JSValue{Boolean: Some(true)}, nil
				}
			}
		}
		return JSValue{Boolean: Some(false)}, nil
	}

	// For string search, cast both to strings
	searchStr, err := castToString(search)
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "contains: cannot cast search to string")
	}

	itemStr, err := castToString(item)
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "contains: cannot cast item to string")
	}

	result := strings.Contains(strings.ToLower(searchStr), strings.ToLower(itemStr))
	return JSValue{Boolean: Some(result)}, nil
}

// equalJSValues compares two JSValues for equality
func equalJSValues(a, b JSValue) bool {
	if a.Type() != b.Type() {
		return false
	}

	switch {
	case a.Null.IsPresent:
		return true
	case a.Undefined.IsPresent:
		return true
	case a.Boolean.IsPresent:
		return a.Boolean.Value == b.Boolean.Value
	case a.String.IsPresent:
		return strings.EqualFold(a.String.Value, b.String.Value)
	case a.Int.IsPresent:
		return a.Int.Value == b.Int.Value
	case a.Float.IsPresent:
		return a.Float.Value == b.Float.Value
	default:
		return false
	}
}

// funcStartsWith implements the startsWith(searchString, searchValue) function.
// Returns true when searchString starts with searchValue.
// This function is not case sensitive. Casts values to a string.
func funcStartsWith(args ...JSValue) (JSValue, error) {
	if len(args) < 2 {
		return JSValue{}, oops.Errorf("startsWith requires 2 arguments, got %d", len(args))
	}

	searchString, err := castToString(args[0])
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "startsWith: cannot cast searchString to string")
	}

	searchValue, err := castToString(args[1])
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "startsWith: cannot cast searchValue to string")
	}

	result := strings.HasPrefix(strings.ToLower(searchString), strings.ToLower(searchValue))
	return JSValue{Boolean: Some(result)}, nil
}

// funcEndsWith implements the endsWith(searchString, searchValue) function.
// Returns true if searchString ends with searchValue.
// This function is not case sensitive. Casts values to a string.
func funcEndsWith(args ...JSValue) (JSValue, error) {
	if len(args) < 2 {
		return JSValue{}, oops.Errorf("endsWith requires 2 arguments, got %d", len(args))
	}

	searchString, err := castToString(args[0])
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "endsWith: cannot cast searchString to string")
	}

	searchValue, err := castToString(args[1])
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "endsWith: cannot cast searchValue to string")
	}

	result := strings.HasSuffix(strings.ToLower(searchString), strings.ToLower(searchValue))
	return JSValue{Boolean: Some(result)}, nil
}

// funcFormat implements the format(string, replaceValue0, replaceValue1, ..., replaceValueN) function.
// Replaces values in the string, with the variable replaceValueN.
// Variables in the string are specified using the {N} syntax, where N is an integer.
// Escape curly braces using double braces.
func funcFormat(args ...JSValue) (JSValue, error) {
	if len(args) < 2 {
		return JSValue{}, oops.Errorf("format requires at least 2 arguments, got %d", len(args))
	}

	formatStr, err := castToString(args[0])
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "format: cannot cast format string to string")
	}

	// First, replace escaped braces with placeholders
	const (
		openBracePlaceholder  = "\x00OPEN_BRACE\x00"
		closeBracePlaceholder = "\x00CLOSE_BRACE\x00"
	)

	result := strings.ReplaceAll(formatStr, "{{", openBracePlaceholder)
	result = strings.ReplaceAll(result, "}}", closeBracePlaceholder)

	// Replace {N} placeholders with corresponding argument values
	for i := 1; i < len(args); i++ {
		placeholder := fmt.Sprintf("{%d}", i-1)
		replacement, err := castToString(args[i])
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "format: cannot cast argument %d to string", i-1)
		}
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	// Restore escaped braces
	result = strings.ReplaceAll(result, openBracePlaceholder, "{")
	result = strings.ReplaceAll(result, closeBracePlaceholder, "}")

	return JSValue{String: Some(result)}, nil
}

// funcJoin implements the join(array, optionalSeparator) function.
// All values in array are concatenated into a string.
// If optionalSeparator is provided, it is inserted between the concatenated values.
// Otherwise, the default separator ',' is used. Casts values to a string.
func funcJoin(args ...JSValue) (JSValue, error) {
	if len(args) < 1 {
		return JSValue{}, oops.Errorf("join requires at least 1 argument, got %d", len(args))
	}

	separator := ","
	if len(args) >= 2 {
		sep, err := castToString(args[1])
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "join: cannot cast separator to string")
		}
		separator = sep
	}

	// Handle string input (treat as single-element)
	if args[0].String.IsPresent {
		return args[0], nil
	}

	if !args[0].Array.IsPresent {
		// If it's not an array, cast to string and return
		str, err := castToString(args[0])
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "join: cannot cast value to string")
		}
		return JSValue{String: Some(str)}, nil
	}

	arr := args[0].Array.Value
	parts := make([]string, 0, len(arr))
	for i, elem := range arr {
		str, err := castToString(elem)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "join: cannot cast element %d to string", i)
		}
		parts = append(parts, str)
	}

	return JSValue{String: Some(strings.Join(parts, separator))}, nil
}

// funcToJSON implements the toJSON(value) function.
// Returns a pretty-print JSON representation of value.
func funcToJSON(args ...JSValue) (JSValue, error) {
	if len(args) < 1 {
		return JSValue{}, oops.Errorf("toJSON requires 1 argument, got %d", len(args))
	}

	jsonBytes, err := json.Marshal(args[0], jsontext.WithIndent("  "))
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "toJSON: failed to marshal value")
	}

	return JSValue{String: Some(string(jsonBytes))}, nil
}

// funcFromJSON implements the fromJSON(value) function.
// Returns a JSON object or JSON data type for value.
func funcFromJSON(args ...JSValue) (JSValue, error) {
	if len(args) < 1 {
		return JSValue{}, oops.Errorf("fromJSON requires 1 argument, got %d", len(args))
	}

	jsonStr, err := castToString(args[0])
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "fromJSON: cannot cast value to string")
	}

	var goValue any
	if err := json.Unmarshal([]byte(jsonStr), &goValue); err != nil {
		return JSValue{}, oops.Wrapf(err, "fromJSON: failed to parse JSON")
	}

	result, err := UnmarshalFromGo(goValue)
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "fromJSON: failed to convert to JSValue")
	}

	return result, nil
}

// hashFilesWorkspace is the base directory for hashFiles.
// It can be overridden for testing or by setting GITHUB_WORKSPACE.
var hashFilesWorkspace = ""

// funcHashFiles implements the hashFiles(path) function.
// Returns a single hash for the set of files that matches the path pattern.
// You can provide a single path pattern or multiple path patterns.
// The path is relative to the GITHUB_WORKSPACE directory.
func funcHashFiles(args ...JSValue) (JSValue, error) {
	if len(args) < 1 {
		return JSValue{}, oops.Errorf("hashFiles requires at least 1 argument, got %d", len(args))
	}

	workspace := hashFilesWorkspace
	if workspace == "" {
		workspace = os.Getenv("GITHUB_WORKSPACE")
	}
	if workspace == "" {
		// Default to current directory if GITHUB_WORKSPACE is not set
		var err error
		workspace, err = os.Getwd()
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "hashFiles: cannot determine workspace directory")
		}
	}

	// Collect all patterns
	var patterns []string
	for i, arg := range args {
		pattern, err := castToString(arg)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "hashFiles: cannot cast pattern %d to string", i)
		}
		patterns = append(patterns, pattern)
	}

	// Find all matching files
	matchedFiles, err := matchFilesWithPatterns(workspace, patterns)
	if err != nil {
		return JSValue{}, oops.Wrapf(err, "hashFiles: error matching files")
	}

	// If no files matched, return empty string
	if len(matchedFiles) == 0 {
		return JSValue{String: Some("")}, nil
	}

	// Sort files for consistent hashing
	sort.Strings(matchedFiles)

	// Calculate SHA-256 hash for each file, then combine
	fileHashes := make([][]byte, 0, len(matchedFiles))
	for _, file := range matchedFiles {
		hash, err := hashFile(file)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "hashFiles: error hashing file %s", file)
		}
		fileHashes = append(fileHashes, hash)
	}

	// Combine all file hashes into a final hash
	finalHasher := sha256.New()
	for _, h := range fileHashes {
		finalHasher.Write(h)
	}
	finalHash := hex.EncodeToString(finalHasher.Sum(nil))

	return JSValue{String: Some(finalHash)}, nil
}

// matchFilesWithPatterns finds all files matching the given glob patterns.
// Patterns starting with ! are exclusion patterns.
func matchFilesWithPatterns(workspace string, patterns []string) ([]string, error) {
	included := make(map[string]bool)
	excluded := make(map[string]bool)

	for _, pattern := range patterns {
		isExclusion := strings.HasPrefix(pattern, "!")
		if isExclusion {
			pattern = pattern[1:]
		}

		// Normalize the pattern - handle leading /
		pattern = strings.TrimPrefix(pattern, "/")

		fullPattern := filepath.Join(workspace, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return nil, oops.Wrapf(err, "invalid glob pattern: %s", pattern)
		}

		// For ** patterns, we need to do a recursive walk
		if strings.Contains(pattern, "**") {
			matches, err = doubleStarGlob(workspace, pattern)
			if err != nil {
				return nil, oops.Wrapf(err, "error in recursive glob: %s", pattern)
			}
		}

		for _, match := range matches {
			// Only include files, not directories
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.IsDir() {
				continue
			}

			if isExclusion {
				excluded[match] = true
			} else {
				included[match] = true
			}
		}
	}

	// Build final list excluding excluded files
	var result []string
	for file := range included {
		if !excluded[file] {
			result = append(result, file)
		}
	}

	return result, nil
}

// doubleStarGlob handles ** glob patterns by walking the directory tree
func doubleStarGlob(workspace, pattern string) ([]string, error) {
	var matches []string

	// Convert ** pattern to a regex
	regexPattern := globToRegex(pattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, oops.Wrapf(err, "invalid pattern regex: %s", regexPattern)
	}

	err = filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if info.IsDir() {
			return nil
		}

		// Get relative path from workspace
		relPath, err := filepath.Rel(workspace, path)
		if err != nil {
			return nil
		}

		// Convert to forward slashes for consistent matching
		relPath = filepath.ToSlash(relPath)

		if re.MatchString(relPath) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, oops.Wrapf(err, "error walking directory")
	}

	return matches, nil
}

// globToRegex converts a glob pattern to a regex pattern
func globToRegex(glob string) string {
	// Normalize path separators
	glob = filepath.ToSlash(glob)

	// Escape regex special characters except * and ?
	var result strings.Builder
	result.WriteString("^")

	i := 0
	for i < len(glob) {
		c := glob[i]
		switch c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				// ** matches any path
				if i+2 < len(glob) && glob[i+2] == '/' {
					result.WriteString("(?:.*/)?")
					i += 3
					continue
				}
				result.WriteString(".*")
				i += 2
				continue
			}
			// * matches anything except /
			result.WriteString("[^/]*")
		case '?':
			result.WriteString("[^/]")
		case '.', '+', '^', '$', '(', ')', '[', ']', '{', '}', '|', '\\':
			result.WriteString("\\")
			result.WriteByte(c)
		default:
			result.WriteByte(c)
		}
		i++
	}

	result.WriteString("$")
	return result.String()
}

// hashFile calculates the SHA-256 hash of a file
func hashFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, oops.Wrapf(err, "cannot open file")
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, oops.Wrapf(err, "cannot read file")
	}

	return h.Sum(nil), nil
}
