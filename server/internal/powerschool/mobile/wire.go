package mobile

import (
	"bytes"
	"encoding/json"
	"strings"
)

// normalize strips Axis transport metadata from a decoded JSON tree:
// {"@nil": true} becomes null and other "@"-prefixed keys (such as @type) are dropped.
func normalize(v any) any {
	switch t := v.(type) {
	case map[string]any:
		if isNilMarker(t["@nil"]) {
			return nil
		}
		out := make(map[string]any, len(t))
		for k, x := range t {
			if !strings.HasPrefix(k, "@") {
				out[k] = normalize(x)
			}
		}
		return out
	case []any:
		for i, x := range t {
			t[i] = normalize(x)
		}
	}
	return v
}

func isNilMarker(v any) bool {
	switch v {
	case true, "true", "1":
		return true
	}
	return false
}

// List is a repeated XML field. The JSON transport sends it as a missing key, null,
// a single value, or an array; all decode to a slice, with null members dropped.
type List[T any] []T

func (l *List[T]) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*l = nil
		return nil
	}
	if b[0] != '[' {
		var v T
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*l = List[T]{v}
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	out := make(List[T], 0, len(raw))
	for _, r := range raw {
		if string(bytes.TrimSpace(r)) == "null" {
			continue
		}
		var v T
		if err := json.Unmarshal(r, &v); err != nil {
			return err
		}
		out = append(out, v)
	}
	*l = out
	return nil
}

// Text is a value the JSON transport may coerce to a string, number, or boolean
// (scores like "--", numeric-looking comments, period names like 1). Valid is false for null.
type Text struct {
	Value string
	Valid bool
}

func (t *Text) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	switch {
	case string(b) == "null":
		*t = Text{}
	case len(b) > 0 && b[0] == '"':
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*t = Text{s, true}
	default:
		*t = Text{string(b), true}
	}
	return nil
}

func (t Text) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(t.Value)
}

// Number is a numeric value the WSDL may declare as a string. Non-numeric text is not a number.
type Number struct {
	Value float64
	Valid bool
}

func (n *Number) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		b = []byte(s)
	}
	var f float64
	if string(b) == "null" || json.Unmarshal(b, &f) != nil {
		*n = Number{}
		return nil
	}
	*n = Number{f, true}
	return nil
}
