package runner

import "strings"

// based on github.com/actions/runner/src/Runner.Common/ActionCommand.cs EscapeMapping

type escapingMapping struct {
	Token, Replacement string
}

var escapingDataMapping = []escapingMapping{
	{Token: "\r", Replacement: "%0D"},
	{Token: "\n", Replacement: "%0A"},
	// notice % is at the end
	{Token: "%", Replacement: "%25"},
}

var escapingPropertyMapping = []escapingMapping{
	{Token: "\r", Replacement: "%0D"},
	{Token: "\n", Replacement: "%0A"},
	{Token: ":", Replacement: "%3A"},
	{Token: ",", Replacement: "%2C"},
	// notice % is at the end
	{Token: "%", Replacement: "%25"},
}

func unescape(mapping []escapingMapping, data string) string {
	for _, mp := range mapping {
		data = strings.ReplaceAll(data, mp.Replacement, mp.Token)
	}
	return data
}

var escapingLegacyMapping = []escapingMapping{
	{Token: ";", Replacement: "%3B"},
	{Token: "\r", Replacement: "%0D"},
	{Token: "\n", Replacement: "%0A"},
	{Token: "]", Replacement: "%5D"},
	// notice % is at the end
	{Token: "%", Replacement: "%25"},
}

func _escape(mapping []escapingMapping, data string) string {
	for _, mp := range mapping {
		data = strings.ReplaceAll(data, mp.Token, mp.Replacement)
	}
	return data
}
