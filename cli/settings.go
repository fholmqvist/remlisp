package cli

type Settings struct {
	Path  string `arg:"positional" help:"path to the input file"`
	Out   string `arg:"-o, --out" help:"path of the output file"`
	REPL  bool   `help:"start REPL"`
	Run   bool   `help:"run the output (deno)"`
	Debug bool   `help:"print debug info"`
}
