package outputs

import (
	"encoding/json"
	"regexp"
	"strings"
)

const outputPrefix = "airplane_output"
const defaultOutputName = "output"

var outputRegexp = regexp.MustCompile((`^airplane_output(?::(?:("[^"]*")|('[^']*')|([^ ]+))?)? (.*)$`))

func IsOutput(s string) bool {
	return strings.HasPrefix(s, outputPrefix)
}

func ParseOutputName(s string) string {
	if matches := outputRegexp.FindStringSubmatch(s); matches != nil {
		var outputName string
		if matches[1] != "" {
			outputName = strings.Trim(matches[1], "\"")
		} else if matches[2] != "" {
			outputName = strings.Trim(matches[2], "'")
		} else if matches[3] != "" {
			outputName = matches[3]
		}
		outputName = strings.TrimSpace(outputName)
		if outputName != "" {
			return outputName
		}
	}
	return defaultOutputName
}

func ParseOutputValue(s string) interface{} {
	var value string
	if matches := outputRegexp.FindStringSubmatch(s); matches != nil {
		value = strings.TrimSpace(matches[4])
	}
	var target interface{}
	if err := json.Unmarshal([]byte(value), &target); err != nil {
		// Interpret this output as a string
		target = value
	}
	return target
}
