package vm

import (
	"fmt"
	"strings"
)

type ErrorWithTrace struct {
	wrapped error
	trace   []TraceFrame
}

func (e *ErrorWithTrace) Unwrap() error {
	return e.wrapped
}

func (e *ErrorWithTrace) Error() string {
	b := new(strings.Builder)
	b.WriteString("error:\n  ")
	b.WriteString(e.wrapped.Error())
	b.WriteString("\ntrace (latest scope last):\n")
	for i, f := range e.trace {
		fmt.Fprintf(b, "%3d: ", i)
		b.WriteString(f.String())
		b.WriteString("\n")
	}
	return b.String()
}

type TraceFrame struct {
	Pos       string // TODO: change to lexer.Position
	ctx       []string
	ctxOffset int
	loc       *int
	args      []string
}

func (t *TraceFrame) String() string {
	b := new(strings.Builder)
	b.WriteString(t.Pos)
	if len(t.ctx) > 0 {
		b.WriteString("\n     context:\n")
		for i, inst := range t.ctx {
			b.WriteString("       ")
			fmt.Fprintf(b, "%3d", i+t.ctxOffset)
			if t.loc != nil {
				fmt.Fprintf(b, " +%03x", i+t.ctxOffset+*t.loc)
			}
			b.WriteString(": ")
			b.WriteString(inst)
			b.WriteString("\n")
		}
	}
	if t.loc != nil {
		b.WriteString("\n     latest loc: ")
		fmt.Fprintf(b, "%03x", *t.loc)
	}
	if len(t.args) > 0 {
		b.WriteString("\n     ")
		b.WriteString(strings.Join(t.args, ", "))
	}
	return b.String()
}
