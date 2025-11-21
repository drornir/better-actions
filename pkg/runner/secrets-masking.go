package runner

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
)

type SecretsMasker struct {
	lock sync.RWMutex

	sensitiveStrings []string
	sensitiveRegexes []*regexp.Regexp
}

func (m *SecretsMasker) AddString(s ...string) {
	var encodersExpanded []string

	for _, str := range s {
		if str == "" {
			continue
		}
		encodersExpanded = append(encodersExpanded, expandMaskedSecretValueWithEncoders(str)...)
	}
	if len(encodersExpanded) == 0 {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.sensitiveStrings = append(m.sensitiveStrings, encodersExpanded...)
	// sort by length, desc. This makes sure we mask bigger values before smaller values, making this safer of there are nested secrets to mask
	slices.SortFunc(m.sensitiveStrings, func(a, b string) int {
		return len(b) - len(a)
	})
	for i := 0; i < len(m.sensitiveStrings)-1; i++ {
		cur := m.sensitiveStrings[i]
		dups := 0
		for cur == m.sensitiveStrings[i+1+dups] {
			dups++
			if i+dups+1 == len(m.sensitiveStrings) {
				break
			}
		}
		if dups > 0 {
			after := m.sensitiveStrings[i+1+dups:]
			m.sensitiveStrings = append(m.sensitiveStrings[:i+1], after...)
		}
	}
}

func (m *SecretsMasker) AddRegex(r ...*regexp.Regexp) {
	if len(r) == 0 {
		return
	}
	m.lock.Lock()
	m.sensitiveRegexes = append(m.sensitiveRegexes, r...)
	m.lock.Unlock()

	reAsStrings := make([]string, 0, len(r))
	for _, re := range r {
		reAsStrings = append(reAsStrings, re.String())
	}
	m.AddString(reAsStrings...)
}

func (m *SecretsMasker) Mask(s string) string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, sens := range m.sensitiveStrings {
		s = strings.ReplaceAll(s, sens, "***")
	}
	for _, senseexp := range m.sensitiveRegexes {
		s = senseexp.ReplaceAllLiteralString(s, "***")
	}

	return s
}

type secretValueEncoder func(string) string

// src/Runner.Common/HostContext.cs
// src/Sdk/DTLogging/Logging/ValueEncoders.cs
var secretValueEncoders = [...]secretValueEncoder{
	// noop
	func(v string) string { return v },
	// Base64StringEscape
	func(v string) string { return encodeAsBase64Partial(v, 0, false) },
	func(v string) string { return encodeAsBase64Partial(v, 0, true) },
	// Base64StringEscapeShift1
	func(v string) string { return encodeAsBase64Partial(v, 1, false) },
	func(v string) string { return encodeAsBase64Partial(v, 1, true) },
	// Base64StringEscapeShift2
	func(v string) string { return encodeAsBase64Partial(v, 2, false) },
	func(v string) string { return encodeAsBase64Partial(v, 2, true) },
	// CommandLineArgumentEscape
	func(v string) string { return strings.ReplaceAll(v, "\"", "\\\"") },
	// ExpressionStringEscape
	func(v string) string { return strings.ReplaceAll(v, "'", "''") },
	// JsonStringEscape
	func(v string) string { return encodeAsJSONString(v, false) },
	func(v string) string { return encodeAsJSONString(v, true) },
	// UriDataEscape
	func(v string) string { return url.QueryEscape(v) },
	func(v string) string { return url.PathEscape(v) },
	// XmlDataEscape
	func(v string) string { return encodeAsXMLString(v) },
	// TrimQuotes (originally it was TrimDoubleQuotes)
	func(v string) string { return strings.Trim(v, "\"'") },
	// PowerShellPreAmpersandEscape
	func(v string) string { return powerShellPreAmpersandEscape(v) },
	// PowerShellPostAmpersandEscape
	func(v string) string { return powerShellPostAmpersandEscape(v) },
}

// encodeAsBase64Partial implements the logic from GH Runner:
// Base64 is 6 bits -> char
// A byte is 8 bits
// When end user doing something like base64(user:password)
// The length of the leading content will cause different base64 encoding result on the password
// So we add base64(value shifted 1 and two bytes) as secret as well.
//
//	   B1         B2      B3                    B4      B5     B6     B7
//	000000|00 0000|0000 00|000000|            000000|00 0000|0000 00|000000|
//	Char1  Char2    Char3   Char4
//
// See the above, the first byte has a character beginning at index 0, the second byte has a character beginning at index 4, the third byte has a character beginning at index 2 and then the pattern repeats
// We register byte offsets for all these possible values
func encodeAsBase64Partial(s string, shift uint, urlencoded bool) string {
	var encoding *base64.Encoding
	if urlencoded {
		encoding = base64.RawURLEncoding // Raw means no padding
	} else {
		encoding = base64.RawStdEncoding // Raw means no padding
	}
	out := bytes.Buffer{}
	encoder := base64.NewEncoder(encoding, &out)
	v := []byte(s)[shift:]

	if _, err := encoder.Write(v); err != nil {
		encodingName := "standard"
		if urlencoded {
			encodingName = "URL"
		}
		panic(fmt.Errorf("base64 encoding with shift=%d and encoding %s failed on write: %w", shift, encodingName, err))
	}
	if err := encoder.Close(); err != nil {
		encodingName := "standard"
		if urlencoded {
			encodingName = "URL"
		}
		panic(fmt.Errorf("base64 encoding with shift=%d and encoding %s failed on close: %w", shift, encodingName, err))
	}

	return out.String()
}

// encodeAsJSONString converts to a JSON string and then remove the leading/trailing double-quote.
func encodeAsJSONString(s string, htmlescape bool) string {
	out := bytes.Buffer{}
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(htmlescape)
	if err := encoder.Encode(s); err != nil {
		panic(fmt.Errorf("JSON encoding failed on write: %w", err))
	}
	return strings.Trim(out.String(), "\"\n")
}

// encodeAsXMLString converts to an XML string
func encodeAsXMLString(s string) string {
	out := bytes.Buffer{}
	if err := xml.EscapeText(&out, []byte(s)); err != nil {
		panic(fmt.Errorf("XML escaping failed: %w", err))
	}
	return out.String()
}

// powerShellPreAmpersandEscape masks the first section of a secret when PowerShell
// splits it on ampersands and causes an error.
//
// If the secret is passed to PS as a command and it causes an error, sections of it
// can be surrounded by color codes or printed individually.
//
// The secret secretpart1&secretpart2&secretpart3 would be split into 2 sections:
// 'secretpart1&secretpart2&' and 'secretpart3'. This method masks for the first section.
//
// The secret secretpart1&+secretpart2&secretpart3 would be split into 2 sections:
// 'secretpart1&+' and (no 's') 'ecretpart2&secretpart3'. This method masks for the first section.
func powerShellPreAmpersandEscape(value string) string {
	if value == "" || !strings.Contains(value, "&") {
		return ""
	}

	var secretSection string
	if strings.Contains(value, "&+") {
		idx := strings.Index(value, "&+")
		secretSection = value[:idx+2] // +2 for "&+" length
	} else {
		idx := strings.LastIndex(value, "&")
		secretSection = value[:idx+1] // +1 for "&" length
	}

	// Don't mask short secrets
	if len(secretSection) < 6 {
		return ""
	}

	return secretSection
}

// powerShellPostAmpersandEscape masks the second section of a secret when PowerShell
// splits it on ampersands and causes an error.
func powerShellPostAmpersandEscape(value string) string {
	if value == "" || !strings.Contains(value, "&") {
		return ""
	}

	var secretSection string
	if strings.Contains(value, "&+") {
		idx := strings.Index(value, "&+")
		// +2 for "&+" length, +1 to skip the letter that got colored
		if len(value) > idx+3 {
			secretSection = value[idx+3:]
		}
	} else {
		idx := strings.LastIndex(value, "&")
		secretSection = value[idx+1:] // +1 to skip the "&"
	}

	// Don't mask short secrets
	if len(secretSection) < 6 {
		return ""
	}

	return secretSection
}

func expandMaskedSecretValueWithEncoders(secretValue string) []string {
	var expandedValues []string
	for _, encoder := range secretValueEncoders {
		encoded := encoder(secretValue)
		if encoded == "" {
			continue
		}
		expandedValues = append(expandedValues, encoded)
	}
	return expandedValues
}
