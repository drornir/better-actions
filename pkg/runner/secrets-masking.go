package runner

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

type secretValueEncoder func(string) string

// src/Runner.Common/HostContext.cs
var secretValueEncoders = [...]secretValueEncoder{
	// noop
	func(v string) string { return v },
	// Base64StringEscape
	func(v string) string { return encodeBase64Partial(v, 0, false) },
	// Base64StringEscapeShift1
	func(v string) string { return encodeBase64Partial(v, 1, false) },
	// Base64StringEscapeShift2
	func(v string) string { return encodeBase64Partial(v, 2, false) },
	// Base64URLStringEscape
	func(v string) string { return encodeBase64Partial(v, 0, true) },
	// Base64URLStringEscapeShift1
	func(v string) string { return encodeBase64Partial(v, 1, true) },
	// Base64URLStringEscapeShift2
	func(v string) string { return encodeBase64Partial(v, 2, true) },
	// CommandLineArgumentEscape
	// ExpressionStringEscape
	// JsonStringEscape
	// UriDataEscape
	// XmlDataEscape
	// TrimDoubleQuotes
	// PowerShellPreAmpersandEscape
	// PowerShellPostAmpersandEscape
}

func encodeBase64Partial(s string, shift uint, urlencoded bool) string {
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
		panic(fmt.Errorf("base64 encoding with shift=%d and encoding %s failed on write: %w", opts.shift, encodingName, err))
	}
	if err := encoder.Close(); err != nil {
		encodingName := "standard"
		if urlencoded {
			encodingName = "URL"
		}
		panic(fmt.Errorf("base64 encoding with shift=%d and encoding %s failed on close: %w", opts.shift, encodingName, err))
	}

	return out.String()
}

func expandMaskedSecretValueWithEncoders(secretValue string) []string {
	var expandedValues []string
	for _, encoder := range secretValueEncoders {
		expandedValues = append(expandedValues, encoder(secretValue))
	}
	return expandedValues
}
