package bot

import "encoding/json"

// ParseJSONInput parses JSON string into a struct
func ParseJSONInput[T any](source string) (T, error) {
	var args T

	if err := json.Unmarshal([]byte(source), &args); err != nil {
		return *(new(T)), err
	}

	return args, nil
}
