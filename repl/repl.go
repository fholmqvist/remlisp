package repl

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"

	h "github.com/fholmqvist/remlisp/highlight"
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
	fmt.Println()
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
		fmt.Println(err)
		return
	}
	if strings.HasPrefix(out, "(exit") {
		done <- true
		return
	}
	if print {
		fmt.Println(out)
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
