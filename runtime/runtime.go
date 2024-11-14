package runtime

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"time"

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
	deno := exec.Command("deno", "run", "--allow-read", "runtime/runtime.mjs")
	stdin, err := deno.StdinPipe()
	if err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	stdout, err := deno.StdoutPipe()
	if err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	stderr, err := deno.StderrPipe()
	if err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	if err := deno.Start(); err != nil {
		return nil, &e.Error{Msg: err.Error()}
	}
	r := &Runtime{
		deno:   deno,
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
	_, stdlib, erre := r.cmp.CompileFile("stdlib/stdlib.rem", false, r.exp)
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
	// TODO: Synchronize?
	time.Sleep(time.Millisecond * 25)
	var out string
	scanner := bufio.NewScanner(r.stdout)
	if scanner.Scan() {
		out += scanner.Text()
	}
	return out, nil
}
