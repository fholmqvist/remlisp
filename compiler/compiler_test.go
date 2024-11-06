package compiler

import (
	"testing"

	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
)

func TestCompiler(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
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
			output: "0.000000",
		},
		{
			input:  "1234.0",
			output: "1234.000000",
		},
		{
			input:  "-1234.0",
			output: "-1234.000000",
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
			input:  ":a",
			output: "\":a\"",
		},
		{
			input:  "+",
			output: "+",
		},
		{
			input:  "(1 2 3 4)",
			output: "(1 2 3 4)",
		},
		{
			input:  "[1 2 3 4]",
			output: "[1 2 3 4]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			code := getCode(t, tt.input)
			if code != tt.output {
				t.Fatalf("\n\nexpected\n\n%s\n\ngot\n\n%s\n\n",
					h.Code(tt.output), h.Code(code))
			}
		})
	}
}

func getCode(t *testing.T, input string) string {
	bb := []byte(input)
	lexer, err := lexer.New(bb)
	if err != nil {
		t.Fatal(err)
	}
	tokens, erre := lexer.Lex()
	if erre != nil {
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), erre.String(bb))
	}
	parser, err := parser.New(tokens)
	if err != nil {
		t.Fatal(err)
	}
	exprs, erre := parser.Parse()
	if erre != nil {
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), erre.String(bb))
	}
	comp, err := New(exprs)
	if err != nil {
		t.Fatal(err)
	}
	code, err := comp.Compile(exprs)
	if err != nil {
		t.Fatal(err)
	}
	return code
}
