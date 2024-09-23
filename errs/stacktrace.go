package errs

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

const (
	defaultStackSkip = 3
)

var (
	content   string
	stackSkip = defaultStackSkip
	traceable bool
)

func SetStackSkip(skip int) {
	stackSkip = skip
}

func SetTraceable(x bool) {
	traceable = x
}

type frame uintptr

func (f frame) pc() uintptr {
	return uintptr(f) - 1
}

func (f frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func (f frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

func (f frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

func (f frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			_, _ = io.WriteString(s, f.name())
			_, _ = io.WriteString(s, "\n\t")
			_, _ = io.WriteString(s, f.file())
		default:
			_, _ = io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		_, _ = io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		_, _ = io.WriteString(s, funcName(f.name()))
	case 'v':
		f.Format(s, 's')
		_, _ = io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

type stackTrace []frame

func (st stackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				if !isOutput(fmt.Sprintf("%+v", f)) {
					continue
				}
				_, _ = io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []frame(st))
		default:
			st.formatSlice(s, verb)
		}
	case 's':
		st.formatSlice(s, verb)
	}
}

func (st stackTrace) formatSlice(s fmt.State, verb rune) {
	_, _ = io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			_, _ = io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	_, _ = io.WriteString(s, "]")
}

func isOutput(str string) bool {
	return strings.Contains(str, content)
}

func callers() stackTrace {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(stackSkip, pcs[:])
	stack := pcs[0:n]
	st := make([]frame, len(stack))
	for i := 0; i < len(st); i++ {
		st[i] = frame((stack)[i])
	}
	return st
}

func funcName(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}
