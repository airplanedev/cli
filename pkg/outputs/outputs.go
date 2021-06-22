package outputs

import (
	"encoding/json"
	"strings"
)

const outputPrefix = "airplane_output"
const outputSeparator = ":"
const defaultOutputName = "output"

func IsOutput(s string) bool {
	return strings.HasPrefix(s, outputPrefix)
}

func ParseOutputName(s string) string {
	nonDefaultPrefix := outputPrefix + outputSeparator
	if strings.HasPrefix(s, nonDefaultPrefix) {
		trimmed := strings.TrimPrefix(s, nonDefaultPrefix)
		outputName := strings.SplitN(trimmed, " ", 2)[0]
		if outputName != "" {
			return outputName
		}
	}
	return defaultOutputName
}

func ParseOutputValue(s string) interface{} {
	var value string
	split := strings.SplitN(s, " ", 2)
	if len(split) >= 2 {
		value = strings.TrimSpace(split[1])
	}
	var target interface{}
	if err := json.Unmarshal([]byte(value), &target); err != nil {
		// Interpret this output as a string
		target = value
	}
	return target
}
