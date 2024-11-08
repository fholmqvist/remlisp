package pp

import (
	"encoding/json"
	"fmt"
	"strings"
)

func FromJS(js []byte) (string, error) {
	if len(js) == 0 {
		return "nil", nil
	}
	var obj any
	if err := json.Unmarshal(js, &obj); err != nil {
		return "", err
	}
	var s strings.Builder
	switch obj := obj.(type) {
	case map[string]any:
		s.WriteByte('{')
		for k, v := range obj {
			s.WriteString(fmt.Sprintf("%s %v", k, v))
		}
		s.WriteByte('}')
		return s.String(), nil
	case []any:
		s.WriteByte('[')
		for i, v := range obj {
			s.WriteString(fmt.Sprintf("%v", v))
			if i < len(obj)-1 {
				s.WriteByte(' ')
			}
		}
		s.WriteByte(']')
		return s.String(), nil
	default:
		return fmt.Sprintf("%v", obj), nil
	}
}
