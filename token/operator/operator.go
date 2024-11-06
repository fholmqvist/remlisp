package operator

import (
	"fmt"

	e "github.com/fholmqvist/remlisp/err"
)

type Operator uint8

const (
	UNKNOWN Operator = iota
	ADD
	SUB
	MUL
	DIV
	MOD
	EQ
	NEQ
	LT
	LTE
	GT
	GTE
)

func From(s string) (Operator, error) {
	switch s {
	case "+":
		return ADD, nil
	case "-":
		return SUB, nil
	case "*":
		return MUL, nil
	case "/":
		return DIV, nil
	case "%":
		return MOD, nil
	case "=":
		return EQ, nil
	case "!=":
		return NEQ, nil
	case "<":
		return LT, nil
	case "<=":
		return LTE, nil
	case ">":
		return GT, nil
	case ">=":
		return GTE, nil
	default:
		return UNKNOWN, fmt.Errorf("unknown operator: %s", s)
	}
}

func (o Operator) String() string {
	switch o {
	case UNKNOWN:
		return "UNKNOWN"
	case ADD:
		return "+"
	case SUB:
		return "-"
	case MUL:
		return "*"
	case DIV:
		return "/"
	case MOD:
		return "%"
	case EQ:
		return "="
	case NEQ:
		return "!="
	case LT:
		return "<"
	case LTE:
		return "<="
	case GT:
		return ">"
	case GTE:
		return ">="
	default:
		e.Panic("unknown operator", fmt.Sprintf("%d", o))
		return ""
	}
}
