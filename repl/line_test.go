package repl

import "testing"

func TestLineCommands(t *testing.T) {
	tests := []struct {
		input    string
		action   func(*line)
		expected string
	}{
		{
			input:    "hello",
			expected: "hello",
		},
		{
			input: "hello",
			action: func(l *line) {
				l.reset()
			},
			expected: "",
		},
		{
			input: "hello",
			action: func(l *line) {
				l.backspace()
				l.backspace()
			},
			expected: "hel",
		},
		{
			input: "hello",
			action: func(l *line) {
				l.left()
				l.left()
				l.delete()
				l.delete()
			},
			expected: "hel",
		},
		{
			input: "hello",
			action: func(l *line) {
				l.left()
				l.right()
				l.backspace()
				l.backspace()
			},
			expected: "hel",
		},
		{
			input: "hello",
			action: func(l *line) {
				l.home()
				l.delete()
			},
			expected: "ello",
		},
		{
			input: "hello",
			action: func(l *line) {
				l.home()
				l.end()
				l.backspace()
				l.backspace()
			},
			expected: "hel",
		},
		{
			input: "",
			action: func(l *line) {
				l.openParen()
			},
			expected: "()",
		},
		{
			input: "",
			action: func(l *line) {
				l.openBracket()
			},
			expected: "[]",
		},
		{
			input: "",
			action: func(l *line) {
				l.openBrace()
			},
			expected: "{}",
		},
		{
			input: "",
			action: func(l *line) {
				l.quotation()
			},
			expected: "\"\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			line := newLine()
			line.add([]byte(tt.input))
			if tt.action != nil {
				tt.action(line)
			}
			result := string(line.get())
			if tt.expected != result {
				t.Fatalf("\n\nexpected\n\n%s\n\ngot\n\n%s\n\n", tt.expected, result)
			}
		})
	}
}
