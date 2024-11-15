package expr

func removeQuotes(ex Expr) Expr {
	switch ex := ex.(type) {
	case *Quasiquote:
		return removeQuotes(ex.E)
	case *Quote:
		return removeQuotes(ex.E)
	case *Unquote:
		return removeQuotes(ex.E)
	case *Vec:
		for i, e := range ex.V {
			ex.V[i] = removeQuotes(e)
		}
		return ex
	case *List:
		for i, e := range ex.V {
			ex.V[i] = removeQuotes(e)
		}
		return ex
	default:
		return ex
	}
}
