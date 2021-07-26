package outputs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOutput(tt *testing.T) {
	for _, test := range []struct {
		name          string
		log           string
		expectedName  string
		expectedValue interface{}
	}{
		{
			name:          "default no colon",
			log:           "airplane_output hello",
			expectedName:  "output",
			expectedValue: "hello",
		},
		{
			name:          "default with colon",
			log:           "airplane_output: true",
			expectedName:  "output",
			expectedValue: true,
		},
		{
			name:          "named",
			log:           "airplane_output:named [1, 2, 3]",
			expectedName:  "named",
			expectedValue: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:          "quoted string",
			log:           "airplane_output \"hello world\"",
			expectedName:  "output",
			expectedValue: "hello world",
		},
		{
			name:          "named with extra spaces",
			log:           "airplane_output:my_output   hello world  ",
			expectedName:  "my_output",
			expectedValue: "hello world",
		},
		{
			name:          "named with tabs",
			log:           "airplane_output:tabs \thello\tworld",
			expectedName:  "tabs",
			expectedValue: "hello\tworld",
		},
		{
			name:          "empty value with colon",
			log:           "airplane_output:",
			expectedName:  "output",
			expectedValue: "",
		},
		{
			name:          "empty value with colon and space",
			log:           "airplane_output: ",
			expectedName:  "output",
			expectedValue: "",
		},
		{
			name:          "empty value no colon",
			log:           "airplane_output",
			expectedName:  "output",
			expectedValue: "",
		},
		{
			name:          "empty value no colon and space",
			log:           "airplane_output ",
			expectedName:  "output",
			expectedValue: "",
		},
		{
			name:          "named and quoted",
			log:           "airplane_output:\"output\" hello",
			expectedName:  "output",
			expectedValue: "hello",
		},
		{
			name:          "named and quoted with spaces",
			log:           "airplane_output:\"output value\" hello",
			expectedName:  "output value",
			expectedValue: "hello",
		},
		{
			name:          "named and quoted with quoted value",
			log:           "airplane_output:\"output\" \"hello\"",
			expectedName:  "output",
			expectedValue: "hello",
		},
		{
			name:          "named and quoted with spaces with quoted value",
			log:           "airplane_output:\"output value\" \"hello\"",
			expectedName:  "output value",
			expectedValue: "hello",
		},
		{
			name:          "empty quoted name",
			log:           "airplane_output:\"\" hello",
			expectedName:  "output",
			expectedValue: "hello",
		},
		{
			name:          "named and single quoted",
			log:           "airplane_output:'output' hello",
			expectedName:  "output",
			expectedValue: "hello",
		},
		{
			name:          "named and single quoted with spaces",
			log:           "airplane_output:'output value' hello",
			expectedName:  "output value",
			expectedValue: "hello",
		},
		{
			name:          "named and single quoted with quoted value",
			log:           "airplane_output:'output' \"hello\"",
			expectedName:  "output",
			expectedValue: "hello",
		},
		{
			name:          "named and single quoted with spaces with quoted value",
			log:           "airplane_output:'output value' \"hello\"",
			expectedName:  "output value",
			expectedValue: "hello",
		},
		{
			name:          "empty single quoted name",
			log:           "airplane_output:'' hello",
			expectedName:  "output",
			expectedValue: "hello",
		},
		{
			name:          "malformed output",
			log:           "airplane_output:''' hello",
			expectedName:  "'''",
			expectedValue: "hello",
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedName, ParseOutputName(test.log))
			require.Equal(t, test.expectedValue, ParseOutputValue(test.log))
		})
	}
}
