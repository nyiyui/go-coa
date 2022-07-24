package vm

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"gitlab.com/coalang/go-coa/try2/compile"
	"gitlab.com/coalang/go-coa/try2/parser"
)

const ctxLength = 9

type Program struct {
	offset int
	insts  []compile.Instruction
}

func NewProgram(insts []compile.Instruction) *Program {
	return &Program{
		insts: insts,
	}
}

type Value interface {
	Evaler() parser.Evaler
}

type VMCallable interface {
	VMCall(v *VM) (Value, error)
	// VMCall is used to call a function using the VM.
	// The args are in v.s().args.
	// The return value should be returned (alongside a nil error).
	// A new scope for this VMCall must be made by the caller, which must be
	// discarded after the call.
}

type valueProxy struct{ evaler parser.Evaler }

func (v *valueProxy) Evaler() parser.Evaler { return v.evaler }

func (v *valueProxy) String() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "proxy %s", v.evaler)
	return b.String()
}

func proxySlice(evalers []parser.Evaler) []Value {
	v := make([]Value, len(evalers))
	for i, e := range evalers {
		v[i] = &valueProxy{e}
	}
	return v
}

func unproxySlice(values []Value) []parser.Evaler {
	v := make([]parser.Evaler, len(values))
	for i, v2 := range values {
		v[i] = v2.Evaler()
	}
	return v
}

type VM struct {
	scopes []*Scope
	// scopes is a stack of scopes.
	// Each scope is equivalent to a frame in a call stack.

	globalProg *Program
	// globalProg is the program to be executed.
	// This is mainly used for debugging.used for debugging, etc
}

// NewVM makes a new blank VM.
func NewVM() *VM {
	return &VM{
		scopes: make([]*Scope, 0),
	}
}

// newScope makes a new Scope for this VM.
func (v *VM) newScope(note string) *Scope {
	return &Scope{
		vars:    make([]Value, 0),
		args:    make([]Value, 0),
		dynvars: parser.NewBase(),
		vm:      v,
		note:    note,
	}
}

// Execute sets prog as the global program to execute.
// This cannot be called concurrently.
func (v *VM) Execute(prog *Program) (err error) {
	v.globalProg = prog
	v.pushScope(v.newScope("Execute"))
	return v.exec(prog)
}

// Scope is equivalent to a frame in the call stack.
type Scope struct {
	sn *scopeSnapshot

	stack []Value
	// stack stores temporary values.

	vars []Value
	// vars stores local variables.

	varNames []string
	// varNames stores the names of local variables (in vars) for debugging.

	args []Value
	// args stores arguments passed to the function (if applicable).
	// aargs is separate from vars (instead of being in the same slice like the JVM (I think))
	// as there is no pre-runtime checking of arguments, and this is easier to implement.

	dynvars map[string]parser.Evaler
	// dynvars stores variables for OpDynVar. Mainly used for builtins.

	pos string
	// pos stores the position of the last OpPos call.

	parent *Scope
	// parent stores the parent scope when inherit()ed.

	vm *VM
	// vm is the VM that this scope belongs to.

	latestInst *compile.Instruction
	// latestInst stores the last instruction executed.

	latestLoc *int
	// latestLoc stores the location of the last instruction executed.

	note string
	// note is for debugging.
}

func (s *Scope) inherit(note string) *Scope {
	return &Scope{
		vars:    make([]Value, 0),
		args:    make([]Value, 0),
		dynvars: s.dynvars,
		pos:     s.pos,
		parent:  s,
		vm:      s.vm,
		note:    note,
	}
}

// s returns the current (bottom-most) scope.
func (v *VM) s() *Scope {
	return v.scopes[len(v.scopes)-1]
}

// sLevel returns the scope at level levels above the current scope.
func (v *VM) sLevel(level int) *Scope {
	log.Println("sLevel i", len(v.scopes)-1-level)
	return v.scopes[len(v.scopes)-1-level]
}

func (v *VM) pushScope(s *Scope) { v.scopes = append(v.scopes, s) }
func (v *VM) popScope() *Scope {
	s := v.scopes[len(v.scopes)-1]
	v.scopes = v.scopes[:len(v.scopes)-1]
	return s
}

func (v *VM) pushFrame(v2 Value) {
	if v2 == nil {
		panic("nil frame")
	}
	v.s().stack = append(v.s().stack, v2)
}
func (v *VM) popFrame() Value {
	v2 := v.s().stack[len(v.s().stack)-1]
	v.s().stack = v.s().stack[:len(v.s().stack)-1]
	return v2
}

func (v *VM) logInst(p *Program, n int, i compile.Instruction) {
	log.Printf("+%03xo +%03xn +%03x: %s", p.offset, n, p.offset+n, &i)
}

func (v *VM) logCurrent() {
	b := new(strings.Builder)
	fmt.Fprintf(b, "scopes:\n")
	for i, s := range v.scopes {
		fmt.Fprintf(b, "  %d: %s\n", i, s.note)
	}

	fmt.Fprintf(b, "current level %d / note %s / parent %p / %s\n", len(v.scopes)-1, v.s().note, v.s().parent, v.s().pos)
	fmt.Fprint(b, "frame stack:\n")
	for i, v2 := range v.s().stack {
		if i == len(v.s().stack)-1 {
			fmt.Fprintf(b, "  c: %s\n", v2)
		} else {
			fmt.Fprintf(b, "  %d: %v\n", i, v2)
		}
	}
	fmt.Fprint(b, "frame vars:\n")
	for i, v2 := range v.s().vars {
		name := v.s().varNames[i]
		fmt.Fprintf(b, "  %d / %s: %v\n", i, name, v2)
	}
	fmt.Fprint(b, "frame args:\n")
	for i, a := range v.s().args {
		fmt.Fprintf(b, "  %d: %v\n", i, a)
	}
	log.Print(b.String())
}

func (v *VM) exec(p *Program) (err error) {
	// v.pushScope(v.newScope())
	for i := 0; i < len(p.insts); i++ {
		inst := p.insts[i]
		v.s().latestInst = &inst
		{
			loc := p.offset + i
			v.s().latestLoc = &loc
		}
		// NOTE: we cannot do things like i += x using for...range
		if inst.Opcode != compile.OpNop &&
			inst.Opcode != compile.OpWrap &&
			inst.Opcode != compile.OpUnwrap {
			v.logInst(p, i, inst)
			v.logCurrent()
		} else {
			log.Printf("+%03x elide", p.offset+i)
		}
		switch inst.Opcode {
		case compile.OpNop:
		case compile.OpPos:
			v.s().pos = inst.B.(string)
		case compile.OpDynVar:
			name := inst.B.(string)
			dynvar := v.s().dynvars[name]
			v.pushFrame(&valueProxy{dynvar})
		case compile.OpWrap:
			log.Printf("┌ %s: %s", v.s().pos, inst.B.(string))
		case compile.OpUnwrap:
			log.Printf("└ %s: %s", v.s().pos, inst.B.(string))

		case compile.OpVarDeclare:
			log.Printf("declared %d on level %d", inst.A, len(v.scopes)-1)
			n := inst.A
			v.s().vars = make([]Value, n)
			v.s().varNames = make([]string, n)

		case compile.OpBool:
			v.pushFrame(&valueProxy{&parser.Bool{Content: inst.A == 1}})

		case compile.OpVarReassign, compile.OpVarAssign:
			v.logInst(p, i, inst)
			log.Println("assign, current vars:", v.s().vars)
			v.s().vars[inst.A] = v.popFrame()
			v.s().varNames[inst.A] = inst.B.(string)
		case compile.OpVarLoad:
			index, level := inst.A, inst.B.(int)
			s := v.s()
			var v2 Value
			var name string
			if s.sn != nil && level != 0 {
				vs := s.sn.get(level, index)
				v2 = vs.v2
				name = vs.name
			} else {
				s := v.sLevel(level)
				v2 = s.vars[inst.A]
				if err := v.nilCheck("loaded nil var", v2); err != nil {
					return err
				}
				name = s.varNames[inst.A]
			}
			v.pushFrame(v2)
			log.Printf("loaded %s: %v", name, v2)
		case compile.OpArgLoad:
			v.logCurrent()
			v.pushFrame(v.s().args[inst.A])

		case compile.OpCall:
			log.Println("======OpCall1======")
			uses := inst.A
			// stack:
			// other things
			// callee
			// arg 1
			// arg 2
			// arg n
			s := v.s()
			baseI := len(s.stack) - uses
			v.logCurrent()
			log.Println("======OpCall2a======", baseI)
			callee := s.stack[baseI]
			args := s.stack[baseI+1 : baseI+uses]
			s.stack = s.stack[:baseI]
			if callee == nil {
				return v.wrapError(errors.New("callee is nil"))
			}
			log.Println("======OpCall2b====== pre", callee)
			callee, err := v.s().eval(callee)
			if err != nil {
				return v.wrapError(err)
			}
			log.Println("======OpCall3====== post", callee)
			v.logCurrent()

			switch callee := callee.(type) {
			case VMCallable:
				var returned Value
				err := func() error {
					v.pushScope(v.s().inherit("VMCall"))
					defer v.popScope()
					// TODO: reset stack
					v.s().args = args
					v.logCurrent()
					log.Println("============VMCall============")
					returned, err = callee.VMCall(v)
					if err != nil {
						return err
					}
					return nil
				}()
				if err != nil {
					return v.wrapError(err)
				}
				v.pushFrame(returned)
			default:
				log.Println("======OpCall3a======", callee)
				callable, ok := callee.Evaler().(parser.Callable)
				if !ok {
					return v.wrapError(fmt.Errorf("%s is not callable", callee))
				}
				log.Printf("calling %s with %s", callee, args)
				result, err := callable.Call(v.s().iEnv(), unproxySlice(args))
				if err != nil {
					return v.wrapError(fmt.Errorf("calling: %w", err))
				}
				v.pushFrame(&valueProxy{result})
			}
			log.Println("======OpCall4======")
			v.logCurrent()

		case compile.OpBlockStart:
			// 1. check validity of block
			endInst := p.insts[i+inst.A-1]
			if endInst.Opcode != compile.OpBlockEnd {
				return v.wrapError(fmt.Errorf("block end not found"))
			}

			innerInsts := p.insts[i+1 : i+inst.A-1]

			// 2. save current scope as immutable (also flatten to 1 outer scope)
			v.logCurrent()
			sn, err := func() (*scopeSnapshot, error) {
				v.pushScope(v.s().inherit("block snap"))
				v.logCurrent()
				defer v.popScope()
				sn, err := v.scopeSnapshot(innerInsts)
				if err != nil {
					return nil, v.wrapError(err)
				}
				log.Println("sn", sn)
				return sn, nil
			}()
			if err != nil {
				return err
			}

			// 3. copy A insts (until Be)
			// NOTE: p.insts includes Os and Oe, strip them off for Instructions
			log.Println("new block")
			v.logCurrent()
			block := Instructions{v.s().pos, i + 1, innerInsts, sn}
			// NOTE: not using the key: value format for struct because this part should define everything in Instructions (at least for now)
			v.pushFrame(&block)
			log.Printf("block %x → %x", i, i+inst.A)
			i += inst.A
			v.logCurrent()
		case compile.OpBlockEnd:
			panic("should be skipped")

		case compile.OpLitNumber:
			b := parser.Number(inst.B.(float64))
			v.pushFrame(&valueProxy{&b})
		case compile.OpLitString:
			b := parser.String{Content: inst.B.(string)}
			v.pushFrame(&valueProxy{&b})
		case compile.OpLitRune:
			b := parser.Rune(inst.B.(rune))
			v.pushFrame(&valueProxy{&b})

		default:
			return fmt.Errorf("unknown opcode: %s", inst.Opcode.Full())
		}
	}
	return nil
}

func (s *Scope) eval(v Value) (Value, error) {
	log.Println("v", v)
	if e := v.Evaler(); e != nil {
		r, err := e.Eval(s.iEnv())
		if err != nil {
			return nil, s.wrapError(err)
		}
		return &valueProxy{r}, nil
	}
	return v, nil
}

func (s *Scope) wrapError(err error) error {
	return fmt.Errorf("%s: %s", s.pos, err)
}

func (v *VM) trace() (trace []TraceFrame) {
	trace = make([]TraceFrame, len(v.scopes))
	for i := len(v.scopes) - 1; i >= 0; i-- {
		s := v.scopes[i]

		args := make([]string, len(s.args))
		for i, a := range s.args {
			args[i] = fmt.Sprint(a)
		}

		loc := s.latestLoc

		var ctx []string
		var ctxOffset int
		if loc != nil {
			ctx = make([]string, ctxLength)
			ctxOffset = -len(ctx) / 2
			for i := range ctx {
				gp := v.globalProg
				j := *loc + i + ctxOffset
				if j >= len(gp.insts) {
					// end of program
					ctx[i] = "past end of program"
					ctx = ctx[:i+1]
					break
				}
				ctx[i] = gp.insts[j].String()
			}
		}

		trace[i] = TraceFrame{
			Pos:       s.pos,
			ctxOffset: ctxOffset,
			ctx:       ctx,
			loc:       loc,
			args:      args,
		}
	}
	return trace
}

func (v *VM) wrapError(err error) *ErrorWithTrace {
	if err2, ok := err.(*ErrorWithTrace); ok {
		err = err2.wrapped // NOTE: update (replace) trace
	}
	return &ErrorWithTrace{wrapped: err, trace: v.trace()}
}

func (v *VM) nilCheck(reason string, v2 Value) error {
	if v2 == nil {
		return v.wrapError(errors.New(reason))
	}
	return nil
}

func (v *VM) scopeSnapshot(insts []compile.Instruction) (*scopeSnapshot, error) {
	sn := new(scopeSnapshot)
	log.Println("snap", insts)
	for _, inst := range insts {
		// look for OpVarLoads with a non-zero level.
		if inst.Opcode != compile.OpVarLoad {
			log.Printf("2ignoring %s", &inst)
			continue
		}
		index, level := inst.A, inst.B.(int)
		if level == 0 {
			log.Printf("ignoring %s", &inst)
			continue
		}
		log.Printf("adding %s", &inst)
		name := v.sLevel(level).varNames[index]
		v2 := v.sLevel(level).vars[index]
		vs := varSnapshot{name: name, v2: v2}
		sn.set(level, index, vs)
		log.Println("snapshot", vs)
	}
	return sn, nil
}
