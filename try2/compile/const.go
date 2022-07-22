package compile

import (
	"fmt"
	"strings"
)

type Instruction struct {
	Opcode Opcode
	A      int
	B      interface{}
}

func (i *Instruction) String() string {
	b := strings.Builder{}
	b.WriteString(i.Opcode.String())
	b.WriteString(fmt.Sprintf("(%d", i.A))
	if i.B != nil {
		b.WriteString(fmt.Sprintf(", %v", i.B))
	}
	b.WriteString(")")
	return b.String()
}

type Instructions []Instruction

type wrap struct {
	Level int
	Name  string
}

func (is Instructions) String() string {
	b := strings.Builder{}
	wraps := make([]wrap, 0)
	for _, i := range is {
		if i.Opcode != OpUnwrap {
			b.WriteString(strings.Repeat("│ ", len(wraps)))
		}
		switch i.Opcode {
		case OpWrap:
			name := i.B.(string)
			w := wrap{i.A, name}
			wraps = append(wraps, w)
			fmt.Fprintf(&b, "┌ %s\n", name)
		case OpUnwrap:
			cw := wraps[len(wraps)-1]
			if cw.Level != i.A {
				b.WriteString("warning: unwrap level mismatch\n")
			}
			wraps = wraps[:len(wraps)-1]
			b.WriteString(strings.Repeat("│ ", len(wraps)))
			b.WriteString("└ end\n")
		case OpPos:
			fmt.Fprintf(&b, "pos %s\n", i.B)
		default:
			b.WriteString(i.String())
			b.WriteString("\n")
		}
	}
	return b.String()
}

type Opcode uint8

func (o Opcode) String() string {
	switch o {
	case OpNop:
		return "nop"
	case OpWrap:
		return "wrap"
	case OpUnwrap:
		return "unwrap"
	case OpConst:
		return "const"
	case OpPop:
		return "pop"
	case OpDynVar:
		return "dynvar"

	case OpWellKnown:
		return "wk"
	case OpBool:
		return "bool"

	case OpCall:
		return "call"
	case OpLit:
		return "L"

	case OpMakeList:
		return "Ml"
	case OpMakeString:
		return "Ms"

	case OpBlockStart:
		return "Ls"
	case OpBlockEnd:
		return "Le"

	case OpBundleStart:
		return "Bs"
	case OpStrandStart:
		return "Ss"
	case OpStrandTodo:
		return "St"
	case OpStrandReverseDeps:
		return "Srd"
	case OpStrandInvoke:
		return "Si"
	case OpStrandEnd:
		return "Se"
	case OpBundleEnd:
		return "Be"

	case OpPos:
		return "pos"

	case OpLitNumber:
		return "Ln"

	default:
		return fmt.Sprintf("invalid %x", o)
	}
}

// NOTE: @ = ttop of stack; @1 = 2nd item on stack

const (
	OpNop Opcode = iota
	OpWrap
	OpUnwrap
	OpConst
	OpPop // pop A frames from stack
	OpDynVar

	OpWellKnown // pushes a well-known value with number A
	OpBool      // pushes true if A == 0, false if A == 1, panics otherwise

	OpCall
	// OpCall calls @ with A arguments (@1, @2, @3, ...)
	// i.e. @(@1, @2, @3, ...)

	OpLit // OpLit pushes a literal value from B. The type is specified using A.

	OpMakeList
	// OpMakeList makes a list with A elements
	// i.e. [@1, @2, @3, ...]

	OpMakeString
	// OpMakeString makes a string with A frames.
	// i.e. [@1, @2, @3, ...].map(string_concat)

	OpBlockStart
	OpBlockEnd

	OpBundleStart
	OpStrandStart
	OpStrandTodo
	OpStrandReverseDeps
	OpStrandInvoke
	OpStrandEnd // noop
	OpBundleEnd // noop

	OpPos // sets the position until the next OpPos with B

	OpLitNumber
)

func op1(code Opcode) Instruction { return Instruction{code, 0, nil} }

func op(code Opcode, a int) Instruction { return Instruction{code, a, nil} }

func op3(code Opcode, a int, b interface{}) Instruction { return Instruction{code, a, b} }
