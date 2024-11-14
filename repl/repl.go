package repl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"

	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/pp"
	"github.com/fholmqvist/remlisp/runtime"
)

const MAX_HISTORY = 256

type Repl struct {
	rt   *runtime.Runtime
	line *line

	signals chan os.Signal
}

func Run(rt *runtime.Runtime, stdlib []byte) {
	r := Repl{
		rt:   rt,
		line: newLine(),
	}
	r.Run(stdlib)
}

func (r *Repl) Run(stdlib []byte) {
	fd, orig, raw := initTermios()
	setTermios(fd, raw)
	defer setTermios(fd, orig)
	r.signals = make(chan os.Signal, 1)
	signal.Notify(r.signals, syscall.SIGINT, syscall.SIGTERM)
	var (
		done     = make(chan bool, 1)
		signaled = false
	)
	go func() {
		<-r.signals
		done <- true
		signaled = true
	}()
	r.evalExprs(stdlib, done, false)
	go func() {
		for {
			input := r.input()
			fmt.Println()
			if len(input) == 0 {
				continue
			}
			r.evalExprs(input, done, true)
		}
	}()
	<-done
	if signaled {
		fmt.Printf("%s\n\nGoodbye!\n\n", h.Code("(exit)"))
	} else {
		fmt.Printf("\nGoodbye!\n\n")
	}
}

func (r *Repl) evalExprs(input []byte, done chan bool, print bool) {
	out, err := r.rt.Eval(input)
	if err != nil {
		fmt.Println(err.String(input) + "\n")
		return
	}
	if strings.HasPrefix(out, "(exit") {
		done <- true
		return
	}
	if print {
		r.printResponse(input, out)
	}
}

func (r *Repl) printResponse(input []byte, out string) {
	fmt.Println()
	defer fmt.Println()
	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		fmt.Println(out)
		return
	}
	if r, ok := result["result"]; ok {
		rstr, ok := r.(string)
		if !ok {
			fmt.Println(r)
			return
		}
		if rstr == `"use strict"` {
			if bytes.Contains(input, []byte("(fn ")) {
				name := bytes.Split(input, []byte("(fn "))[1]
				name = name[:bytes.Index(name, []byte(" "))]
				fmt.Println(h.Code(fmt.Sprintf("<fn %s>", name)))
			} else if bytes.Contains(input, []byte("(macro ")) {
				name := bytes.Split(input, []byte("(macro "))[1]
				name = name[:bytes.Index(name, []byte(" "))]
				fmt.Println(h.Code(fmt.Sprintf("<macro %s>", name)))
			}
			return
		}
		pretty, err := pp.FromJS([]byte(rstr))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(h.Code(pretty))
	} else {
		errstr, ok := result["error"]
		if !ok {
			fmt.Println(h.Code("nil"))
			return
		}
		fmt.Println(h.Bold(h.Red(errstr.(string))))
	}
}

func initTermios() (int, syscall.Termios, syscall.Termios) {
	fd := int(os.Stdin.Fd())
	orig := getTermios(fd)
	raw := orig
	raw.Lflag &^= syscall.ICANON | syscall.ECHO
	return fd, orig, raw
}

func getTermios(fd int) syscall.Termios {
	var termios syscall.Termios
	syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
	return termios
}

func setTermios(fd int, termios syscall.Termios) {
	syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
}
