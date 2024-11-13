package runtime

import (
	"bufio"
	"io"
	"log"
	"os/exec"

	"github.com/fholmqvist/remlisp/compiler"
	e "github.com/fholmqvist/remlisp/err"
	"github.com/fholmqvist/remlisp/expander"
)

type Runtime struct {
	deno   *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	cmp *compiler.Compiler
	exp *expander.Expander
}

func New(cmp *compiler.Compiler, exp *expander.Expander) (*Runtime, *e.Error) {
	nodejs := exec.Command("node", "runtime/node.mjs")
	stdin, err := nodejs.StdinPipe()
	if err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	stdout, err := nodejs.StdoutPipe()
	if err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	stderr, err := nodejs.StderrPipe()
	if err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	if err := nodejs.Start(); err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	r := &Runtime{
		deno:   nodejs,
		stdin:  stdin,
		stdout: stdout,
		cmp:    cmp,
		exp:    exp,
	}
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("stderr: %s", scanner.Text())
		}
	}()
	_, stdlib, erre := r.cmp.CompileFile("stdlib/stdlib.rem", false, nil)
	if erre != nil {
		return nil, erre
	}
	if _, erre = r.Send(stdlib); erre != nil {
		return nil, erre
	}
	return r, nil
}

func (r *Runtime) Eval(input []byte) (string, *e.Error) {
	cmp, err := r.cmp.Compile(input, r.exp)
	if err != nil {
		return "", err
	}
	return r.Send(cmp)
}

func (r *Runtime) Send(s string) (string, *e.Error) {
	if _, err := r.stdin.Write([]byte(s)); err != nil {
		return "", &e.Error{Msg: err.Error()}
	}
	var out string
	scanner := bufio.NewScanner(r.stdout)
	if scanner.Scan() {
		out += scanner.Text()
	}
	return out, nil
}
