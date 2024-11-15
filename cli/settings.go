package cli

type Settings struct {
	Path  string `arg:"positional"`
	REPL  bool   `repl:"start REPL"`
	Debug bool   `help:"print debug info"`
}
