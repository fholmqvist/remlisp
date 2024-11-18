package stdlib

import (
	_ "embed"
)

//go:embed stdfns.rem
var StdFns []byte

//go:embed stdmacros.rem
var StdMacros []byte
