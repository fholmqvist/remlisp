package expander

import (
	"strings"
	"testing"

	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
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
	exprs, erre = New(exprs).Expand()
	if erre != nil {
		t.Fatal(erre)
	}
	var s strings.Builder
	for _, expr := range exprs {
		s.WriteString(expr.String())
	}
	return s.String()
}
