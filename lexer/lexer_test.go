package lexer

import (
	"testing"

	h "github.com/fholmqvist/remlisp/highlight"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input:  "nil",
			output: "nil",
		},
		{
			input:  "0",
			output: "0",
		},
		{
			input:  "1234",
			output: "1234",
		},
		{
			input:  "-1234",
			output: "-1234",
		},
		{
			input:  "0.0",
			output: "0.00",
		},
		{
			input:  "1234.0",
			output: "1234.00",
		},
		{
			input:  "-1234.0",
			output: "-1234.00",
		},
		{
			input:  "true",
			output: "true",
		},
		{
			input:  "false",
			output: "false",
		},
		{
			input:  "example_identifier",
			output: "example_identifier",
		},
		{
			input:  "\"example_string\"",
			output: "\"example_string\"",
		},
		{
			input:  ":atom",
			output: ":atom",
		},
		{
			input:  ";; comments are ignored",
			output: "",
		},
		{input: "(", output: "("},
		{input: ")", output: ")"},
		{input: "[", output: "["},
		{input: "]", output: "]"},
		{input: " ", output: " "},
		{input: ",", output: ","},
		{input: "+", output: "+"},
		{input: "-", output: "-"},
		{input: "*", output: "*"},
		{input: "/", output: "/"},
		{input: "%", output: "%"},
		{input: " 1", output: "1"},
		{input: ".", output: "."},
		{input: "&", output: "&"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			bb := []byte(tt.input)
			lexer, err := New(bb)
			if err != nil {
				t.Fatal(err)
			}
			tokens, erre := lexer.Lex()
			if erre != nil {
				t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), erre.String(bb))
			}
			if len(tokens) > 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if len(tokens) == 0 {
				return
			}
			tk := tokens[0].String()
			if tk != tt.output {
				t.Fatalf("\n\nexpected\n\n%s\n\ngot\n\n%s\n\n",
					h.Code(tt.output), h.Code(tk))
			}
		})
	}
}
