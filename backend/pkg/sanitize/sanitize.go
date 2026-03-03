package sanitize

import (
	"encoding/json"

	"github.com/microcosm-cc/bluemonday"
)

// strict policy strips all HTML tags and attributes.
var strict = bluemonday.StrictPolicy()

// Text sanitizes a plain text string by stripping all HTML tags.
func Text(input string) string {
	return strict.Sanitize(input)
}

// JSON sanitizes all string values in a JSON string recursively.
// Non-string values (numbers, booleans, nulls) are preserved as-is.
// Returns the original string if it is not valid JSON.
func JSON(input string) string {
	if input == "" || input == "{}" || input == "[]" {
		return input
	}

	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return Text(input)
	}

	sanitized := sanitizeValue(data)

	result, err := json.Marshal(sanitized)
	if err != nil {
		return Text(input)
	}
	return string(result)
}

func sanitizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return strict.Sanitize(val)
	case map[string]interface{}:
		for k, v := range val {
			val[k] = sanitizeValue(v)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = sanitizeValue(v)
		}
		return val
	default:
		return v
	}
}
