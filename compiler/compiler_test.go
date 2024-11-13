package compiler

import (
	"strings"
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
			input:  "(+ 1 1 1)",
			output: "(1 + 1 + 1)",
		},
		{
			input:  "(- 1 1 1)",
			output: "(1 - 1 - 1)",
		},
		{
			input:  "(* 1 1 1)",
			output: "(1 * 1 * 1)",
		},
		{
			input:  "(/ 1 1 1)",
			output: "(1 / 1 / 1)",
		},
		{
			input:  "(add 1 1)",
			output: "add(1, 1);",
		},
		{
			input:  "(1 2 3 4)",
			output: "[1, 2, 3, 4]",
		},
		{
			input:  "[1 2 3 4]",
			output: "[1, 2, 3, 4]",
		},
		{
			input:  "(fn add [x y] (+ x y))",
			output: "const add = (x, y) => (x + y)",
		},
		{
			input:  "(fn id-array [& x] x)",
			output: "const id_array = (...x) => x",
		},
		{
			input:  "(fn pair->sum [[x y]] (+ x y))",
			output: "const pair_arrow_sum = ([x, y]) => (x + y)",
		},
		{
			input:  "(. (Array 10) (fill 1) (map (fn [_ i] i)))",
			output: "Array(10).fill(1).map((_, i) => i)",
		},
		{
			input:  "(if (< 1 2) 1 2)",
			output: "(() => (1 < 2) ? 1 : 2)()",
		},
		{
			input:  "(do 1 2 3)",
			output: "(() => { 1; 2; return 3; })();",
		},
		{
			input:  "(var x 1)",
			output: "let x = 1;",
		},
		{
			input:  "(set x 2)",
			output: "x = 2;",
		},
		{
			input:  "(get [1 2] 0)",
			output: "[1, 2][0]",
		},
		{
			input:  "{:a 1}",
			output: "({\":a\": 1})",
		},
		{
			input:  "(while (< 1 2) (println \"infinite loop!\"))",
			output: "(() => { while ((1 < 2)) { println(\"infinite loop!\"); } })();",
		},
		{
			input:  "(macro inc [n] (+ n 1))",
			output: "// (macro inc [n] (+ n 1))",
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
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), erre.String(bb))
	}
	parser := parser.New()
	exprs, erre := parser.Parse(tokens)
	if erre != nil {
		t.Fatalf("\n\n%s:\n\n%v\n\n", h.Bold("error"), erre.String(bb))
	}
	comp := New()
	code, err := comp.Compile(exprs)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(code)
}
