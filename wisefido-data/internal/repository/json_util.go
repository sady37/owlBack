package repository

import "encoding/json"

func jsonRawOrString(s string) any {
	if s == "" {
		return s
	}
	var tmp any
	if err := json.Unmarshal([]byte(s), &tmp); err == nil {
		return json.RawMessage([]byte(s))
	}
	return s
}




