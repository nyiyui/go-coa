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
	for n, i := range is {
		fmt.Fprintf(&b, "+%03x: ", n)
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
			if len(wraps) == 0 {
				b.WriteString(strings.Repeat("│ ", len(wraps)))
				fmt.Fprintf(&b, "warning: unwrap without wrap: %s\n", &i)
				continue
			}
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

type OpcodeInfo struct {
	Name  string
	Short string
}

var opcodeInfo = [...]*OpcodeInfo{
	OpNop:    {"Nop", "n"},
	OpWrap:   {"Wrap", "w"},
	OpUnwrap: {"Unwrap", "u"},
	OpPop:    {"Pop", "p"},
	OpDynVar: {"DynVar", "d"},

	//OpWellKnown: {"WellKnown", "wk"},
	OpBool: {"Bool", "b"},

	OpVarDeclare:  {"VarDeclare", "Ve"},
	OpVarReassign: {"VarReassign", "Vr"},
	OpVarAssign:   {"VarAssign", "Va"},
	OpVarLoad:     {"VarLoad", "Vl"},
	OpArgLoad:     {"ArgLoad", "Al"},

	OpCall: {"Call", "c"},
	OpLit:  {"Lit", "l"},

	OpMakeList:   {"MakeList", "Ml"},
	OpMakeString: {"MakeString", "Ms"},

	OpBlockStart: {"BlockStart", "Os"},
	OpBlockEnd:   {"BlockEnd", "Oe"},

	OpBundleStart:       {"BundleStart", "Bs"},
	OpStrandStart:       {"StrandStart", "Ss"},
	OpStrandTodo:        {"StrandTodo", "St"},
	OpStrandReverseDeps: {"StrandReverseDeps", "Srd"},
	OpStrandInvoke:      {"StrandInvoke", "Si"},
	OpStrandEnd:         {"StrandEnd", "Se"},
	OpBundleEnd:         {"BundleEnd", "Be"},

	OpPos: {"Pos", "pos"},

	OpLitNumber: {"LitNumber", "Ln"},
	OpLitString: {"LitString", "Ls"},
	OpLitRune:   {"LitRune", "Lr"},
}

func (o Opcode) info() *OpcodeInfo {
	if o >= Opcode(len(opcodeInfo)) {
		return nil
	}
	return opcodeInfo[o]
}

func (o Opcode) String() string {
	info := o.info()
	if info == nil {
		return fmt.Sprintf("I%x", uint8(o))
	}
	return info.Short
}

func (o Opcode) Name() string {
	info := o.info()
	if info == nil {
		return fmt.Sprintf("Invalid_%x", uint8(o))
	}
	return info.Name
}

func (o Opcode) Full() string {
	info := o.info()
	if info == nil {
		return fmt.Sprintf("%x I%x Invalid_%x", uint8(o), uint8(o), uint8(o))
	}
	return fmt.Sprintf("%x %s %s", uint8(o), info.Short, info.Name)
}

// NOTE: @ = ttop of stack; @1 = 2nd item on stack

const (
	OpNop Opcode = iota
	OpWrap
	OpUnwrap
	OpPop // pop A frames from stack
	OpDynVar

	//OpWellKnown // pushes a well-known value with number A
	OpBool // pushes true if A == 0, false if A == 1, panics otherwise

	OpVarDeclare  // declare number of variable used in block onwards and it may reset all vars.
	OpVarReassign // reassigns the top of the stack to be A
	OpVarAssign   // assigns variable at A with @
	OpVarLoad     // load variable at A
	OpArgLoad     // load argument at A

	OpCall
	// OpCall calls @ with A arguments
	// i.e. @(@1, @2, @3, ...)
	//

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
	OpLitString
	OpLitRune
)

func op1(code Opcode) Instruction { return Instruction{code, 0, nil} }

func op(code Opcode, a int) Instruction { return Instruction{code, a, nil} }

func op3(code Opcode, a int, b interface{}) Instruction { return Instruction{code, a, b} }
