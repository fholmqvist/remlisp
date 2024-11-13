package cli

import (
	"fmt"
	"strings"

	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/pp"
	tk "github.com/fholmqvist/remlisp/token"
)

func printLogo() {
	fmt.Println(h.Bold(h.Blue(`_______________________  ___
___  __ \__  ____/__   |/  /
__  /_/ /_  __/  __  /|_/ /
_  _, _/_  /___  _  /  / /
/_/ |_| /_____/  /_/  /_/
				`)))
	{
		fmt.Print(h.Bold(h.Blue("Version: ")))
		fmt.Println(h.Bold(h.Green("0.1.0")))
	}
}

func prettyPrintTokens(tokens []tk.Token) {
	fmt.Printf("%s\n", h.Bold("TOKENS ============="))
	if len(tokens) > 0 {
		for i, t := range tokens {
			num := fmt.Sprintf("%.4d", i)
			fmt.Printf("%s | %s (%T)\n",
				h.Gray(num), h.Code(t.String()), t)
		}
	} else {
		fmt.Println("<no tokens>")
	}
	printLine()
}

func prettyPrintExprs(exprs []ex.Expr) {
	fmt.Printf("%s\n", h.Bold("EXPRESSIONS ========"))
	if len(exprs) > 0 {
		for i, e := range exprs {
			num := fmt.Sprintf("%.4d", i)
			fmt.Printf("%s | %s (%T)\n",
				h.Gray(num), h.Code(e.String()), e)
		}
	} else {
		fmt.Println("<no expressions>")
	}
	printLine()
}

func printExpanderHeader() {
	fmt.Printf("%s\n", h.Bold("EXPANDER ==========="))
}

func prettyPrintCode(code string) {
	fmt.Printf("%s\n", h.Bold("CODE ==============="))
	if len(code) > 0 {
		for i, line := range strings.Split(code, "\n") {
			num := fmt.Sprintf("%.4d", i)
			fmt.Printf("%s | %s\n",
				h.Gray(num), h.Code(line))
		}
	} else {
		fmt.Println("<no code>")
	}
	printLine()
}

func prettyPrintResult(bb []byte) {
	fmt.Printf("%s\n", h.Bold("RESULT ============="))
	if lisp, err := pp.FromJS(bb); err == nil {
		fmt.Printf("%s\n\n", strings.TrimSpace(h.Code(lisp)))
	} else {
		fmt.Printf("%s\n\n", strings.TrimSpace(h.Code(string(bb))))
	}
}

func printLine() {
	fmt.Printf("%s\n\n",
		"====================")
}
