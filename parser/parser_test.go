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
		{
			input:  "%",
			output: "%",
		},
		{
			input:  "=",
			output: "=",
		},
		{
			input:  "!=",
			output: "!=",
		},
		{
			input:  "<=",
			output: "<=",
		},
		{
			input:  ">",
			output: ">",
		},
		{
			input:  ">=",
			output: ">=",
		},
		{
			input:  "()",
			output: "()",
		},
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
		{
			input:  "(while (< 1 2) (println \"infinite loop!\"))",
			output: "(while (< 1 2) (println \"infinite loop!\"))",
		},
		{
			input:  "'(set x 2)",
			output: "'(set x 2)",
		},
		{
			input:  "`(set x 2)",
			output: "`(set x 2)",
		},
		{
			input:  ",1",
			output: ",1",
		},
		{
			input:  "(macro inc [n] (+ n 1))",
			output: "(macro inc [n] (+ n 1))",
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
		{
			input: "(",
			output: &e.Error{
				Start: 0,
				End:   1,
				Msg:   "unexpected end of input",
			},
		},
		{
			input: "(1",
			output: &e.Error{
				Start: 1,
				End:   2,
				Msg:   "unexpected end of input",
			},
		},
		{
			input: "(fn)",
			output: &e.Error{
				Start: 3,
				End:   4,
				Msg:   "expected identifier",
			},
		},
		{
			input: "(fn add)",
			output: &e.Error{
				Start: 7,
				End:   8,
				Msg:   "expected parameters",
			},
		},
		{
			input: "(fn add add)",
			output: &e.Error{
				Start: 11,
				End:   12,
				Msg:   "expected parameters",
			},
		},
		{
			input: "(fn add [])",
			output: &e.Error{
				Start: 10,
				End:   11,
				Msg:   "expected body",
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
	lexer := lexer.New()
	tokens, erre := lexer.Lex(bb)
	if erre != nil {
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), erre.String(bb))
	}
	parser := New()
	exprs, erre := parser.Parse(tokens)
	return exprs, erre
}
