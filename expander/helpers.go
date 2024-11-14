package expander

import (
	"fmt"

	er "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
)

func (e *Expander) inQuasiquote() bool {
	return len(e.quasi) > 0
}

func (e *Expander) pushQuasi() {
	e.quasi = append(e.quasi, struct{}{})
}

func (e *Expander) popQuasi() {
	e.quasi = e.quasi[:len(e.quasi)-1]
}

func (e *Expander) logMacro(m *ex.Macro) {
	num := fmt.Sprintf("%.4d", e.printouts)
	line := fmt.Sprintf("%s: %v", h.Bold("Read macro"), m.Name)
	fmt.Printf("%s | %s\n", h.Gray(num), line)
	e.printouts++
}

func (e *Expander) logMacroExpansion(name string) {
	if e.print {
		num := fmt.Sprintf("%.4d", e.printouts)
		line := fmt.Sprintf("%s: %v", h.Bold("Expanded"), name)
		fmt.Printf("%s | %s\n", h.Gray(num), line)
		e.printouts++
	}
}

func errFromStr(format string, args ...any) *er.Error {
	return &er.Error{Msg: fmt.Sprintf(format, args...)}
}
