package runner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

var envCommandBlockedVariables = map[string]struct{}{
	"NODE_OPTIONS": {},
}

func parseEnvFile(path string) (map[string]string, error) {
	if path == "" {
		return nil, nil
	}

	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, nil
	}

	return parseEnvFileContent(string(content))
}

func parseEnvFileContent(content string) (map[string]string, error) {
	parser := envFileParser{input: content}
	pairs := make(map[string]string)

	for {
		line, _, ok := parser.readLine()
		if !ok {
			break
		}
		if line == "" {
			continue
		}

		equalsIdx := strings.Index(line, "=")
		heredocIdx := strings.Index(line, "<<")

		switch {
		case equalsIdx >= 0 && (heredocIdx < 0 || equalsIdx < heredocIdx):
			key := line[:equalsIdx]
			if key == "" {
				return nil, fmt.Errorf("invalid format %q: name must not be empty", line)
			}
			value := line[equalsIdx+1:]
			if _, blocked := envCommandBlockedVariables[key]; blocked {
				return nil, fmt.Errorf("can't store %s output parameter using '$GITHUB_ENV' command", key)
			}
			pairs[key] = value

		case heredocIdx >= 0 && (equalsIdx < 0 || heredocIdx < equalsIdx):
			parts := strings.SplitN(line, "<<", 2)
			key := parts[0]
			delimiter := parts[1]
			if key == "" || delimiter == "" {
				return nil, fmt.Errorf("invalid format %q: name and delimiter must not be empty", line)
			}
			value, err := parser.readHereDoc(delimiter)
			if err != nil {
				return nil, err
			}
			if _, blocked := envCommandBlockedVariables[key]; blocked {
				return nil, fmt.Errorf("can't store %s output parameter using '$GITHUB_ENV' command", key)
			}
			pairs[key] = value

		default:
			return nil, fmt.Errorf("invalid format %q", line)
		}
	}

	if len(pairs) == 0 {
		return nil, nil
	}

	return pairs, nil
}

type envFileParser struct {
	input string
	index int
}

func (p *envFileParser) readLine() (line string, newline string, ok bool) {
	if p.index >= len(p.input) {
		return "", "", false
	}

	start := p.index
	for p.index < len(p.input) {
		ch := p.input[p.index]
		if ch == '\n' {
			end := p.index
			newline := "\n"
			if end > start && p.input[end-1] == '\r' {
				end--
				newline = "\r\n"
			}
			line = p.input[start:end]
			p.index++
			return line, newline, true
		}
		p.index++
	}

	line = p.input[start:]
	return line, "", true
}

func (p *envFileParser) readHereDoc(delimiter string) (string, error) {
	var builder strings.Builder
	var lastNewline string
	var sawContent bool

	for {
		line, newline, ok := p.readLine()
		if !ok {
			return "", fmt.Errorf("invalid value: matching delimiter not found %q", delimiter)
		}
		if line == delimiter {
			value := builder.String()
			if sawContent && lastNewline != "" && strings.HasSuffix(value, lastNewline) {
				value = value[:len(value)-len(lastNewline)]
			}
			return value, nil
		}
		if newline == "" {
			return "", fmt.Errorf("invalid value: matching delimiter not found %q", delimiter)
		}

		sawContent = true
		builder.WriteString(line)
		builder.WriteString(newline)
		lastNewline = newline
	}
}
