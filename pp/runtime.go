package pp

import (
	"bytes"
	"encoding/json"
	"fmt"

	h "github.com/fholmqvist/remlisp/highlight"
)

func ParseResponse(input []byte, out string) string {
	res, err := ParseResponseRaw(input, out)
	if err != nil {
		return h.Bold(h.Red(err.Error()))
	}
	return h.Code(res)
}

func ParseResponseRaw(input []byte, out string) (string, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return "", err
	}
	if r, ok := result["result"]; ok {
		rstr, ok := r.(string)
		if !ok {
			return fmt.Sprintf("%s", r), nil
		}
		if rstr == `"use strict"` {
			if bytes.Contains(input, []byte("(fn ")) {
				name := bytes.Split(input, []byte("(fn "))[1]
				name = name[:bytes.Index(name, []byte(" "))]
				return h.Code(fmt.Sprintf("<fn %s>", name)), nil
			} else if bytes.Contains(input, []byte("(macro ")) {
				name := bytes.Split(input, []byte("(macro "))[1]
				name = name[:bytes.Index(name, []byte(" "))]
				return h.Code(fmt.Sprintf("<macro %s>", name)), nil
			}
		}
		pretty, err := FromJS([]byte(rstr))
		if err != nil {
			return "", err
		}
		return pretty, nil
	} else {
		errstr, ok := result["error"]
		if !ok {
			return "nil", nil
		}
		return "", fmt.Errorf(errstr.(string))
	}
}
