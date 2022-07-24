package compile

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/parser"
)

const compileBundle = false

type CompileEnv struct {
	Pos       lexer.Position
	constants []parser.Node
}

func NewEnv(pos lexer.Position) *CompileEnv {
	return &CompileEnv{Pos: pos}
}

type Scope struct {
	c              *CompileEnv
	pos            lexer.Position
	keys           []string
	parent         *Scope
	levelsFromRoot int
}

func (c *CompileEnv) newScope(pos lexer.Position) *Scope {
	return &Scope{c: c, pos: pos}
}

func (c *CompileEnv) NewScope() *Scope { return &Scope{c: c} }

func (s *Scope) inherit(pos lexer.Position) *Scope {
	return &Scope{
		c:              s.c,
		pos:            pos,
		parent:         s,
		levelsFromRoot: s.levelsFromRoot + 1,
	}
}

func (s *Scope) wrap(insts []Instruction, name string) []Instruction {
	insts = append([]Instruction{
		op3(OpWrap, s.levelsFromRoot, name),
	}, insts...)
	insts = append(insts,
		op3(OpUnwrap, s.levelsFromRoot, name),
	)
	return insts
}

func (s *Scope) allKeys() []string {
	keys := make([]string, 0, len(s.keys))
	for s != nil {
		keys = append(keys, s.keys...)
		s = s.parent
	}
	return keys
}

func (s *Scope) CompileNodes(n parser.Nodes) ([]Instruction, error) {
	cn, err := s.compileNodes(n)
	if err != nil {
		return nil, err
	}
	return cn.insts(), nil
}

func (s *Scope) compileNodes(n parser.Nodes) (compiledNode, error) {
	nodes := make([]compiledNode, 1, len(n.Content))
	for _, n := range n.Content {
		node, err := s.compileNode(n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	nodes[0] = &instsNode{raw: []Instruction{op(OpVarDeclare, len(s.keys))}}
	return &compiledNodes{Nodes: nodes}, nil
}

func (s *Scope) compileNode(n parser.Node) (compiledNode, error) {
	switch {
	case n.Number != nil:
		return &litNumberNode{Number: float64(*n.Number)}, nil
	case n.ID != nil:
		return s.compileID(n.ID)
	case n.String_ != nil:
		return &litStringNode{String: n.String_.Content}, nil
	case n.Rune != nil:
		return &litRuneNode{Rune: rune(*n.Rune)}, nil
	case n.Call != nil:
		return s.compileCall(n.Call)
	case n.Block != nil:
		return s.compileBlock(n.Block)
	case n.List != nil:
		return s.compileList(n.List)
	default:
		panic("unknown node type")
	}
}

func (s *Scope) compileID(n *parser.ID) (compiledNode, error) {
	switch n.Content {
	case "@true":
		return &booleanNode{Value: true}, nil
	case "@false":
		return &booleanNode{Value: false}, nil
	//case "@include":
	//	return &wellKnownNode{Number: wellKnownInclude}, nil
	default:
		name := n.Content
		if name[0] == '@' {
			return &dynVarNode{Name: name}, nil
		}
		if name[0] == '$' {
			n2, err := strconv.Atoi(name[1:])
			if err != nil {
				return nil, NewError(n.Pos, fmt.Errorf("argument variable must be a number: %w", err))
			}
			return &instsNode{raw: []Instruction{
				op(OpArgLoad, n2),
			}}, nil
		}
		//if name[0] == '@' {
		//	panic(fmt.Sprintf("builtin %s invalid or not implemented", name))
		//}
		//panic("variables not implemented")
		sym, level, ok := s.getSymbol(name)
		if !ok {
			return nil, NewError(n.Pos, errors.New("undefined symbol "+name))
		}
		return &instsNode{raw: s.wrap([]Instruction{
			op3(OpVarLoad, sym, level),
		}, "load "+name)}, nil
		// return &dynVarNode{Name: name}, nil
	}
}

func (c *CompileEnv) registerConstant(n parser.Node) int {
	c.constants = append(c.constants, n)
	return len(c.constants) - 1
}

// Symbol represents a compiled key (ident).
type Symbol = int

func (s *Scope) getSymbol(key string) (sym Symbol, level int, ok bool) {
	level = 0
	for s != nil {
		for i, k := range s.keys {
			if k == key {
				return Symbol(i), level, true
			}
		}
		s = s.parent
		level++
	}
	return 0, 0, false
}

func (s *Scope) compileCall(n *parser.Call) (*instsNode, error) {
	{
		a := n.Content.Content[0]
		if a.ID == nil {
			goto Normal
		}
		define := a.ID.Content == "@def"
		redefine := a.ID.Content == "@mod"
		if !define && !redefine {
			goto Normal
		}
		b := n.Content.Content[1]
		if b.ID == nil {
			return nil, NewError(n.Pos, errors.New("def expects a name"))
		}
		name := (*b.ID).Content
		s.keys = append(s.keys, name)
		sym, level, ok := s.getSymbol(name)
		if !ok {
			panic("symbol not found although it should have been added right before")
		}
		if level != 0 {
			panic("symbol not at level 0 although it should have been defined right before at the level 0")
		}
		value := n.Content.Content[2]
		compiledValue, err := s.compileNode(value)
		if err != nil {
			return nil, NewError(n.Pos, err)
		}
		insts := []Instruction{op3(OpPos, 0, n.Pos.String())}
		insts = append(insts, compiledValue.insts()...)
		if define {
			insts = append(insts, op3(OpVarAssign, sym, name))
		} else if redefine {
			insts = append(insts, op3(OpVarReassign, sym, name))
		}
		return &instsNode{raw: insts}, nil
	}
Normal:
	insts := make([]Instruction, 1, len(n.Content.Content)+1)
	insts[0] = op3(OpPos, 0, n.Pos.String())
	for _, node := range n.Content.Content {
		//for i := len(n.Content.Content) - 1; i >= 0; i-- {
		//	node := n.Content.Content[i]
		compiled, err := s.compileNode(node)
		if err != nil {
			return nil, NewError(node.Pos, err)
		}
		insts = append(insts, compiled.insts()...)
	}
	insts = append(insts, op(OpCall, len(n.Content.Content)))
	return &instsNode{raw: s.wrap(insts, "call "+n.String())}, nil
}

func (s *Scope) compileBlock(n *parser.Block) (compiledNode, error) {
	ns := s.inherit(n.Pos)
	evalers := make([]parser.Evaler, len(n.Content.Content))
	nodes := make([]compiledNode, len(n.Content.Content))
	for i, n := range n.Content.Content {
		evalers[i] = n.Select()
		var err error
		nodes[i], err = ns.compileNode(n)
		if err != nil {
			return nil, err
		}
	}
	ss, err := parser.CompileEvalers(ns.allKeys(), evalers)
	if err != nil {
		return nil, err
	}
	insts := make([]Instruction, 3)
	if compileBundle {
		insts = append(insts, s.compileBundle(n.Pos, ss, nodes)...)
	} else {
		for _, node := range nodes {
			insts = append(insts, node.insts()...)
		}
	}
	insts = append(insts, op1(OpBlockEnd))
	insts[0] = op3(OpPos, 0, n.Pos.String())
	insts[1] = op(OpBlockStart, len(insts)-1)
	keysLen := len(ns.keys)
	if keysLen > 0 {
		insts[2] = op(OpVarDeclare, len(ns.keys))
	} else {
		insts[2] = op1(OpNop)
	}
	return &instsNode{raw: s.wrap(insts, "block")}, nil
}

type bundleNode struct {
	ss    []*parser.Strand
	nodes []compiledNode
}

// insts generates instructions for running this bundle in parallel.
//     OpBundleStart 21
//     OpStrandStart 7
//     OpStrandTodo 2
//     {(@def a 1)}
//     {(@def b a)}
//     OpStrandReverseDeps 2
//     OpStrandInvoke 1
//     OpStrandInvoke 2
//     OpStrandEnd
//     OpStrandStart 5
//     OpStrandTodo 1
//     {(@def c b)}
//     OpStrandReverseDeps 2
//     OpStrandInvoke 2
//     OpStrandEnd
//     OpStrandStart 4
//     OpStrandTodo 1
//     {(@def d [a b c])}
//     OpStrandReverseDeps 0
//     OpStrandEnd
//     OpBundleEnd
//     OpBundleEnd
func (s *Scope) compileBundle(pos lexer.Position, ss []*parser.Strand, nodes []compiledNode) []Instruction {
	insts := make([]Instruction, 3)
	insts[0] = op3(OpWrap, s.levelsFromRoot, "bundle")
	insts[1] = op3(OpPos, 0, pos.String())
	insts[2] = op(OpBundleStart, -1) // placeholder
	for _, ss := range ss {
		insts2 := make([]Instruction, 1)
		insts2[0] = op(OpStrandStart, -1)

		// Strand Todo
		todoInsts := make([]Instruction, 1)
		todoInsts[0] = op(OpStrandTodo, -1)
		for _, nodeI := range ss.Todo {
			node := nodes[nodeI]
			nodeInsts := node.insts()
			todoInsts = append(insts2, nodeInsts...)
		}
		todoInsts = append(todoInsts, op(OpStrandEnd, 0))
		todoInsts[0] = op(OpStrandTodo, len(todoInsts)-1)
		todoInsts = s.wrap(todoInsts, "strand todo")
		insts2 = append(insts2, todoInsts...)

		// Strand Reverse Deps
		rdInsts := make([]Instruction, 1)
		rdInsts[0] = op(OpStrandReverseDeps, -1)
		for _, depI := range ss.ReverseDeps {
			rdInsts = append(rdInsts, op(OpStrandInvoke, depI))
		}
		rdInsts[0] = op(OpStrandReverseDeps, len(rdInsts)-1)
		rdInsts = s.wrap(rdInsts, "strand reverse deps")
		insts2 = append(insts2, rdInsts...)

		insts2[0] = op(OpStrandStart, len(insts2)-1)
		insts2 = append(insts2, op1(OpStrandEnd))
		insts2 = s.wrap(insts2, "strand")
		insts = append(insts, insts2...)
	}
	insts[2] = op(OpBundleStart, len(insts)-1)
	insts = append(insts, op1(OpBundleEnd))
	return s.wrap(insts, "bundle")
}

func (s *Scope) compileList(l *parser.List) (compiledNode, error) {
	insts := make([]Instruction, 0, len(l.Content.Content)+2) // assume each node makes 1+ insts, same some growing operations
	insts = append(insts, op3(OpPos, 0, l.Pos))
	for _, n := range l.Content.Content {
		compiled, err := s.compileNode(n)
		if err != nil {
			return nil, NewError(n.Pos, err)
		}
		insts = append(insts, compiled.insts()...)
		// @ = n
	}
	insts = append(insts, op(OpMakeList, len(l.Content.Content)))
	return &instsNode{raw: s.wrap(insts, "list")}, nil
}

type compiledNode interface {
	insts() []Instruction
	//msgpack.CustomEncoder
	//msgpack.CustomDecoder
}

type compiledNodes struct {
	Nodes []compiledNode
}

func (c *compiledNodes) insts() []Instruction {
	insts := make([]Instruction, 0, len(c.Nodes)) // compiled nodes probably have 1+ nodes, so save some growing
	for _, n := range c.Nodes {
		insts = append(insts, n.insts()...)
	}
	return insts
}

type booleanNode struct {
	Value bool
}

func (b *booleanNode) insts() []Instruction {
	return []Instruction{
		op(OpBool, boolToInt(b.Value)),
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

type instsNode struct {
	raw []Instruction
}

func (i *instsNode) insts() []Instruction {
	return i.raw
}

/*
type wellKnownNumber int

type wellKnownNode struct {
	Number wellKnownNumber
}

func (w *wellKnownNode) insts() []Instruction { return []Instruction{op(OpWellKnown, int(w.Number))} }

const (
	wellKnownInclude wellKnownNumber = iota
)
*/

type litNumberNode struct {
	Number float64
}

func (l *litNumberNode) insts() []Instruction {
	return []Instruction{
		op3(OpLitNumber, 0, l.Number),
	}
}

type litStringNode struct {
	String string
}

func (l *litStringNode) insts() []Instruction {
	return []Instruction{
		op3(OpLitString, 0, l.String),
	}
}

type litRuneNode struct {
	Rune rune
}

func (l *litRuneNode) insts() []Instruction {
	return []Instruction{
		op3(OpLitRune, 0, l.Rune),
	}
}

type dynVarNode struct {
	Name string
}

func (d *dynVarNode) insts() []Instruction {
	return []Instruction{
		op3(OpDynVar, 0, d.Name),
	}
}
