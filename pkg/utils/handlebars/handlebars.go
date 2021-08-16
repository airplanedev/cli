// Light wrapper around handlebars.
// Temporary replacement until we switch to full JS evaluation.
package handlebars

import "github.com/aymerick/raymond"

// Render is an Airplane-customized handlebars renderer. Primarily, it assumes we're interpolating
// for non-html contexts like run parameters and does not perform HTML-escaping.
func Render(source string, ctx map[string]interface{}) (string, error) {
	// Because we're not using this for HTML, treat all string values as safe.
	vals := map[string]interface{}{}
	for k, v := range ctx {
		if s, ok := v.(string); ok {
			vals[k] = raymond.SafeString(s)
		} else {
			vals[k] = v
		}
	}
	return raymond.Render(source, vals)
}
