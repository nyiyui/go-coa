package parser

import (
	"encoding/json"
	"strconv"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/util"
)

type List struct {
	Pos     lexer.Position
	O       string `parser:"@OBrack"`
	Content Nodes  `parser:"@@"`
	C       string `parser:"@CBrack"`
}

func (l *List) Accept(v Visitor) { v.VisitList(l) }

func (l *List) Len() int { return l.Content.Len() }

func (l *List) Index(i int) (key, value Evaler) { return l.Content.Index(i) }

func NewList(content []string) *List {
	content2 := make([]Node, len(content))
	for i, sub := range content {
		content2[i] = toNode(NewString(sub))
	}
	return &List{Content: Nodes{Content: content2}}
}

func (l *List) UnmarshalJSON(bytes []byte) error {
	panic("implement me")
}

func (l *List) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Content.Content)
}

var _ Evaler = new(List)
var _ HasNodes = new(List)
var _ Iter = new(List)

func (l *List) Info(env IEnv) util.Info { return l.Content.Info(env) }
func (l *List) Eval(env IEnv) (_ Evaler, err error) {
	if l.isMap() {
		return newMapFromNodes(l.Content)
	}
	var evaler Evaler
	l2 := &List{
		Pos:     l.Pos,
		Content: Nodes{Content: make([]Node, len(l.Content.Content))},
	}
	for i, node := range l.Content.Content {
		evaler, err = Eval(node.Select(), env)
		if err != nil {
			return nil, err
		}
		l2.Content.Content[i] = Node{Evaler: evaler}
	}
	return l2, nil
}
func (l *List) String() string {
	if l.isMap() {
		return "[m" + l.Content.String() + "]" // TODO: format [mkey value\nkey value] for maps
	}
	if l.isString() {
		return l.string()
	}
	return "[" + l.Content.Inspect() + "]"
}
func (l *List) Inspect() string {
	if l.isMap() {
		return "[m" + l.Content.Inspect() + "]"
	}
	if l.isString() {
		return strconv.Quote(l.string())
	}
	return "[" + l.Content.Inspect() + "]"
}
func (l *List) IDUses() []string     { return l.Content.IDUses() }
func (l *List) IDSets() []string     { return l.Content.IDSets() }
func (l *List) isMap() bool          { return len(l.O) == 2 }
func (l *List) Nodes() Nodes         { return l.Content }
func (l *List) BecomeString() string { return l.string() }
func (l *List) isString() bool {
	for _, node := range l.Content.Content {
		if node.Rune == nil {
			return false
		}
	}
	return true
}
func (l *List) string() string {
	re := ""
	for _, node := range l.Content.Content {
		re += string(*node.Rune)
	}
	return re
}
