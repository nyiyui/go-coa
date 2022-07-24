package parser

import (
	"encoding/json"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Nodes struct {
	_                util.NoCopy
	Pos              lexer.Position
	Content          []Node `parser:"@@*"`
	DisallowParallel bool
	iterI            int
}

func (n *Nodes) Accept(v Visitor) {
	v.VisitNodes(n)
}

func (n *Nodes) Len() int { return len(n.Content) }

func (n *Nodes) Index(i int) (key, value Evaler) { return NewNumber(float64(i)), n.Content[i].Select() }

var _ Evaler = new(Nodes)
var _ Iter = new(Nodes)

func (n *Nodes) Info(env IEnv) util.Info {
	resources := make([]util.ResourceDef, 0)
	for _, node := range n.Content {
		resources = append(resources, node.Select().Info(env).Resources...)
	}
	return util.Info{Resources: resources}
}
func (n *Nodes) Eval(env IEnv) (Evaler, error) {
	if len(n.Content) == 0 {
		return nil, nil
	}
	evalers, err := eval(env, !n.DisallowParallel, n.Select())
	if err != nil {
		return nil, err
	}
	return evalers[len(evalers)-1], nil
}
func (n *Nodes) String() string {
	re := ""
	for _, node := range n.Content {
		re += node.String() + " "
	}
	if len(n.Content) != 0 {
		re = re[:len(re)-1]
	}
	if len(re) > util.MultilineThreshold {
		re = ""
		for i, node := range n.Content {
			switch i {
			case 0:
				re += node.String() + "\n"
			default:
				re += util.Indent(node.String()) + "\n"
			}
			re = re[:len(re)-1]
		}
	}
	return re
}
func (n *Nodes) Inspect() string {
	re := ""
	for _, node := range n.Content {
		re += node.Inspect() + " "
	}
	if len(n.Content) != 0 {
		re = re[:len(re)-1]
	}
	if len(re) > util.MultilineThreshold {
		re = ""
		for i, node := range n.Content {
			switch i {
			case 0:
				re += node.Inspect() + "\n"
			default:
				re += util.Indent(node.Inspect()) + "\n"
			}
		}
		re = re[:len(re)-1]
	}
	return re
}
func (n *Nodes) Select() []Evaler {
	re := make([]Evaler, len(n.Content))
	for i, node := range n.Content {
		re[i] = node.Select()
	}
	return re
}
func (n *Nodes) IDUses() []string {
	re := make([]string, 0)
	for _, node := range n.Content {
		re = append(re, node.Select().IDUses()...)
	}
	return re
}
func (n *Nodes) IDSets() []string {
	re := make([]string, 0)
	for _, node := range n.Content {
		re = append(re, node.Select().IDSets()...)
	}
	return re
}
func (n *Nodes) Pos_() lexer.Position { return n.Pos }

type Node struct {
	Pos     lexer.Position
	Number  *Number `parser:" @Number"`
	ID      *ID     `parser:"|@ID"`
	String_ *String `parser:"|@String"`
	Rune    *Rune   `parser:"|@Rune"`
	Call    *Call   `parser:"|@@"`
	Block   *Block  `parser:"|@@"`
	List    *List   `parser:"|@@"`
	Evaler  Evaler
}

func (n *Node) Accept(v Visitor) { v.VisitNode(n) }

func (n Node) MarshalJSON() ([]byte, error) {
	return n.Select().(json.Marshaler).MarshalJSON()
}

func (n Node) Inspect() string { return util.ToInspect(n.Select()) }
func (n Node) String() string  { return util.ToString(n.Select()) }
func (n Node) Select() Evaler {
	switch {
	case n.Number != nil:
		return n.Number
	case n.ID != nil:
		return n.ID
	case n.String_ != nil:
		return n.String_
	case n.Rune != nil:
		return n.Rune
	case n.Call != nil:
		return n.Call
	case n.Block != nil:
		return n.Block
	case n.List != nil:
		return n.List
	case n.Evaler != nil:
		return n.Evaler
	default:
		return nil
	}
}
func (n Node) Pos_() lexer.Position { return n.Pos }
