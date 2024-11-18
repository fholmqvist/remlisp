package runtime

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"time"

	e "github.com/fholmqvist/remlisp/err"
)

type Runtime struct {
	deno   *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func New(stdlib string) (*Runtime, *e.Error) {
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
	}
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("stderr: %s", scanner.Text())
		}
	}()
	if _, erre := r.Send(stdlib); erre != nil {
		return nil, erre
	}
	return r, nil
}

func (r *Runtime) Send(js string) (string, *e.Error) {
	return r.SendByte([]byte(js))
}

func (r *Runtime) SendByte(js []byte) (string, *e.Error) {
	if _, err := r.stdin.Write(js); err != nil {
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
