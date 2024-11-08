package highlight

import "testing"

func TestHighlights(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    Bold("1"),
			expected: "\x1b[1m1\x1b[0m",
		},
		{
			input:    Blue("1"),
			expected: "\x1b[38;2;0;170;255m1\x1b[0m",
		},
		{
			input:    Purple("1"),
			expected: "\x1b[38;2;64;32;255m1\x1b[0m",
		},
		{
			input:    Green("1"),
			expected: "\x1b[38;2;0;255;136m1\x1b[0m",
		},
		{
			input:    Yellow("1"),
			expected: "\x1b[38;2;255;160;80m1\x1b[0m",
		},
		{
			input:    Red("1"),
			expected: "\x1b[38;2;255;0;80m1\x1b[0m",
		},
		{
			input:    Gray("1"),
			expected: "\x1b[38;2;8;8;8m1\x1b[0m",
		},
		{
			input:    ErrorLine("1"),
			expected: "\x1b[4:3;58;2;255;0;80m1\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if tt.expected != tt.input {
				t.Fatalf("\n\nexpected\n\n%s\n\ngot\n\n%s\n\n", tt.expected, tt.input)
			}
		})
	}
}

func TestCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    Code("()"),
			expected: Blue("(") + Blue(")"),
		},
		{
			input:    ErrorCode("()"),
			expected: ErrorLine(Blue("(")) + ErrorLine(Blue(")")),
		},
		{
			input:    Code("1"),
			expected: Green("1"),
		},
		{
			input:    ErrorCode("1"),
			expected: ErrorLine(Green("1")),
		},
		{
			input:    Code("\"hello\""),
			expected: Green("\"hello\""),
		},
		{
			input:    ErrorCode("\"hello\""),
			expected: ErrorLine(Green("\"hello\"")),
		},
		{
			input:    Code("'hello'"),
			expected: Green("'hello'"),
		},
		{
			input:    ErrorCode("'hello'"),
			expected: ErrorLine(Green("'hello'")),
		},
		{
			input:    Code("hello"),
			expected: "hello",
		},
		{
			input:    ErrorCode("hello"),
			expected: ErrorLine("hello"),
		},
		{
			input:    Code("fn"),
			expected: Purple("fn"),
		},
		{
			input:    ErrorCode("fn"),
			expected: ErrorLine(Purple("fn")),
		},
		{
			input:    Code("+"),
			expected: Blue("+"),
		},
		{
			input:    ErrorCode("+"),
			expected: ErrorLine(Blue("+")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if tt.expected != tt.input {
				t.Fatalf("\n\nexpected\n\n%s\n\ngot\n\n%s\n\n", tt.expected, tt.input)
			}
		})
	}
}
