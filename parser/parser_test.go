package parser

import (
	"strings"
	"testing"

	e "github.com/fholmqvist/remlisp/err"
	"github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
)

func TestParse(t *testing.T) {
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
			input:  ":a",
			output: ":a",
		},
		{input: "%", output: "%"},
		{input: "=", output: "="},
		{input: "!=", output: "!="},
		{input: "<=", output: "<="},
		{input: ">", output: ">"},
		{input: ">=", output: ">="},
		{
			input:  "(1 2 3 4)",
			output: "(1 2 3 4)",
		},
		{
			input:  "[1 2 3 4]",
			output: "[1 2 3 4]",
		},
		{
			input:  "{:a 2 :b 4}",
			output: "{:a 2 :b 4}",
		},
		{
			input:  "(fn add [x y] (+ x y))",
			output: "(fn add [x y] (+ x y))",
		},
		{
			input:  "(fn id-array [& x] x)",
			output: "(fn id-array [& x] x)",
		},
		{
			input:  "(. (Array 10) (fill 1) (map (fn [_ i] i)))",
			output: "(. (Array 10) (fill 1) (map (fn [_ i] i)))",
		},
		{
			input:  "(if (< 1 2) 1 2)",
			output: "(if (< 1 2) 1 2)",
		},
		{
			input:  "(do 1 2 3)",
			output: "(do 1 2 3)",
		},
		{
			input:  "(var x 1)",
			output: "(var x 1)",
		},
		{
			input:  "(set x 2)",
			output: "(set x 2)",
		},
		{
			input:  "(get {:a 1} :a)",
			output: "(get {:a 1} :a)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			exprs, err := getExprs(t, tt.input)
			if err != nil {
				t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), err.String([]byte(tt.input)))
			}
			if len(exprs) > 1 {
				t.Fatalf("expected 1 expr, got %d", len(exprs))
			}
			e := exprs[0].String()
			if e != tt.output {
				t.Fatalf("\n\nexpected\n\n%s\n\ngot\n\n%s\n\n",
					h.Code(tt.output), h.Code(e))
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		input  string
		output *e.Error
	}{
		{
			input: ")",
			output: &e.Error{
				Start: 0,
				End:   1,
				Msg:   "unexpected token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := getExprs(t, tt.input)
			if err == nil {
				t.Fatalf(h.Bold(h.Red("\n\nexpected error, got nil\n")))
			}
			if !errEq(err, tt.output) {
				t.Fatalf("\n\nexpected\n\n%v\n\ngot\n\n%v\n\n",
					tt.output, err)
			}
		})
	}
}

func errEq(a, b *e.Error) bool {
	return a.Start == b.Start &&
		a.End == b.End &&
		strings.Contains(a.Msg, b.Msg)
}

func getExprs(t *testing.T, input string) ([]expr.Expr, *e.Error) {
	bb := []byte(input)
	lexer, err := lexer.New(bb)
	if err != nil {
		t.Fatal(err)
	}
	tokens, erre := lexer.Lex()
	if erre != nil {
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), erre.String(bb))
	}
	parser, err := New(tokens)
	if err != nil {
		t.Fatal(err)
	}
	exprs, erre := parser.Parse()
	return exprs, erre
}
