package expander

import (
	"strings"
	"testing"

	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	compiler "github.com/fholmqvist/remlisp/transpiler"
)

func TestExpander(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input:  "'(1 2 3)",
			output: "(1 2 3)",
		},
		{
			input:  "`(1 2 3)",
			output: "(1 2 3)",
		},
		{
			input:  "`,(1 2 3)",
			output: "[1 2 3]",
		},
		{
			input:  "`,(+ 1 1)",
			output: "2",
		},
		{
			// Do nothing
			input:  "(fn add-one [n] (+ n 1))",
			output: "(fn add-one [n] (+ n 1))",
		},
		{
			input:  "(macro inc [n] `(+ ,n 1)) (var x 0) (inc x)",
			output: "(macro inc [n] `(+ ,n 1)) (var x 0) (+ x 1)",
		},
		{
			input:  "(macro inc-two [[x y]] `[(+ ,x 1) (+ ,y 1)]) (inc-two [1 4])",
			output: "(macro inc-two [[x y]] `[(+ ,x 1) (+ ,y 1)]) [(+ 1 1) (+ 4 1)]",
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
	lexer := lexer.New()
	tokens, erre := lexer.Lex(bb)
	if erre != nil {
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("lexing error"), erre.String(bb))
	}
	parser := parser.New()
	exprs, erre := parser.Parse(tokens)
	if erre != nil {
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("parse error"), erre.String(bb))
	}
	exprs, erre = New(lexer, parser, compiler.New()).Expand(exprs, false)
	if erre != nil {
		t.Fatal(erre)
	}
	var s strings.Builder
	for i, expr := range exprs {
		s.WriteString(expr.String())
		if i < len(exprs)-1 {
			s.WriteByte(' ')
		}
	}
	return s.String()
}
