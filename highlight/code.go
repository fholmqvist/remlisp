package highlight

import (
	"strconv"
	"strings"
)

func Code(str string) string {
	return code(str, false)
}

func ErrorCode(str string) string {
	return code(str, true)
}

func code(str string, errorColor bool) string {
	var (
		s           strings.Builder
		i           = 0
		isString    bool
		isLeadingWS = true
	)
	for i < len(str) {
		switch str[i] {
		case '(', ')', '[', ']', '{', '}':
			isLeadingWS = false
			if errorColor {
				s.WriteString(ErrorLine(Blue(string(str[i]))))
			} else {
				s.WriteString(Blue(string(str[i])))
			}
			i++
		case '"':
			var s2 strings.Builder
			s2.WriteByte('"')
			i++
			for i < len(str) && str[i] != '"' {
				s2.WriteByte(str[i])
				i++
			}
			s2.WriteByte('"')
			s.WriteString(Green(s2.String()))
			i++
		case '\'':
			var s2 strings.Builder
			s2.WriteByte('"')
			i++
			for i < len(str) && str[i] != '\'' {
				s2.WriteByte(str[i])
				i++
			}
			s2.WriteByte('"')
			s.WriteString(Green(s2.String()))
			i++
		default:
			var s2 strings.Builder
			for i < len(str) && !isDelimiter(str[i]) {
				s2.WriteString(string(str[i]))
				i++
			}
			st := s2.String()
			if st != "" {
				isLeadingWS = false
				if isString || isPurple(st) {
					if errorColor {
						s.WriteString(ErrorLine(Purple(st)))
					} else {
						s.WriteString(Purple(st))
					}
				} else if isBlue(st) {
					if errorColor {
						s.WriteString(ErrorLine(Blue(st)))
					} else {
						s.WriteString(Blue(st))
					}
				} else if isGreen(st) {
					if errorColor {
						s.WriteString(ErrorLine(Green(st)))
					} else {
						s.WriteString(Green(st))
					}
				} else {
					if errorColor {
						s.WriteString(ErrorLine(string(st)))
					} else {
						s.WriteString(string(st))
					}
				}
				continue
			} else {
				if isLeadingWS {
					s.WriteString(string(str[i]))
				} else if errorColor {
					s.WriteString(ErrorLine(string(str[i])))
				} else {
					s.WriteString(string(str[i]))
				}
				i++
			}
		}
	}
	return s.String()
}

func isPurple(s string) bool {
	s = strings.TrimSpace(s)
	switch s {
	case "fn", "if", "cond", "case", "match", "while", "break", "continue", "var":
		return true
	default:
		return false
	}
}

func isBlue(s string) bool {
	s = strings.TrimSpace(s)
	switch s {
	case ":=", "=", "+", "-", "*", "/", "%", "!", "!=", "<", ">", "<=", ">=":
		return true
	default:
		return false
	}
}

func isGreen(s string) bool {
	s = strings.TrimSpace(s)
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	switch s {
	case "true", "false", "nil":
		return true
	default:
		return false
	}
}

func isDelimiter(b byte) bool {
	switch b {
	case '(', ')', '[', ']', '{', '}', ',', ';', ' ', '\n', '\t':
		return true
	default:
		return false
	}
}
