package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Number float64

var _ BecomesNumberLike = (*Number)(nil)

var _ Evaler = new(Number)

func (n *Number) Accept(v Visitor) { v.VisitNumber(n) }

func NewNumber(f float64) *Number { n := Number(f); return &n }

func (n *Number) BecomeNumberLike() NumberLike {
	f := Float(*n)
	return &f
}

func (n *Number) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(*n))
}

func (n *Number) UnmarshalJSON(bytes []byte) error {
	var f float64
	err := json.Unmarshal(bytes, &f)
	if err != nil {
		return err
	}
	*n = Number(f)
	return nil
}

func (n *Number) Info(_ *Env) util.Info              { return util.InfoPure }
func (n *Number) Eval(_ *Env, _ int) (Evaler, error) { return n, nil }
func (n *Number) String() string                     { return strconv.FormatFloat(float64(*n), 'f', -1, 64) }
func (n *Number) Inspect() string                    { return n.String() }
func (n *Number) IDUses() []string                   { return nil }
func (n *Number) IDSets() []string                   { return nil }
func (n *Number) BecomeString() string               { return strconv.FormatFloat(float64(*n), 'f', -1, 64) }
func (n *Number) BecomeFloat64() float64             { return float64(*n) }

type ID struct {
	Pos     lexer.Position
	Content string
}

var _ Evaler = new(ID)

func (i *ID) Accept(v Visitor) { v.VisitID(i) }

func (i *ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

func (i *ID) Eval(env *Env, _ int) (Evaler, error) {
	evaler, ok := env.get(i.Content)
	if !ok {
		return nil, fmt.Errorf("%s (%p): %s not found", i.Pos, i, i.Content)
	}
	return evaler, nil
}

func (i *ID) Info(env *Env) util.Info {
	evaler, ok := env.get(i.Content)
	if !ok {
		return util.InfoPure
	}
	return evaler.Info(env)
}

func (i *ID) String() string                { return i.Content }
func (i *ID) Inspect() string               { return i.String() }
func (i *ID) IDUses() []string              { return []string{i.Content} }
func (i *ID) IDSets() []string              { return nil }
func (i *ID) Pos_() lexer.Position          { return i.Pos }
func (i *ID) Capture(values []string) error { i.Content = values[0]; return nil }

var stringTmplPattern = regexp.MustCompile(`\$` + ident)

type String struct {
	Pos     lexer.Position
	Content string
}

func (s *String) Accept(v Visitor) { v.VisitString(s) }

func (s *String) UnmarshalJSON(data []byte) error {
	var s2 string
	err := json.Unmarshal(data, &s2)
	if err != nil {
		return err
	}
	s.Content = s2
	return nil
}

func (s *String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Content)
}

func NewString(s string) *String        { return &String{Content: s} }
func (s *String) Info(_ *Env) util.Info { return util.InfoPure }
func (s *String) Eval(env *Env, order int) (result Evaler, err error) {
	var errOuter error
	s2 := stringTmplPattern.ReplaceAllStringFunc(s.Content, func(match string) string {
		s3 := ID{Content: match[1:]}
		replaced, err := (&s3).Eval(env, order)
		if err != nil {
			errOuter = err
		}
		if replaced == nil {
			return "<nil>"
		}
		return replaced.String()
	})
	if errOuter != nil {
		return nil, errOuter
	}
	return NewString(s2), nil
}
func (s *String) String() string  { return s.Content }
func (s *String) Inspect() string { return strconv.Quote(s.Content) }
func (s *String) IDUses() []string {
	if len(s.Content) == 0 {
		return nil
	}
	matches := stringTmplPattern.FindAllIndex([]byte(s.Content), -1)
	uses := make([]string, len(matches))
	for i, match := range matches {
		uses[i] = s.Content[match[0]+1 : match[1]]
	}
	return uses
}
func (s *String) IDSets() []string     { return nil }
func (s *String) Pos_() lexer.Position { return s.Pos }
func (s *String) Capture(values []string) error {
	unquoted, err := strconv.Unquote(values[0])
	if err != nil {
		return err
	}
	s.Content = unquoted
	return nil
}
func (s *String) Nodes() Nodes {
	nodes := Nodes{Pos: s.Pos, Content: []Node{}}
	for _, r := range s.Content {
		r2 := Rune(r)
		nodes.Content = append(nodes.Content, Node{Rune: &r2})
	}
	return nodes
}
func (s *String) BecomeString() string {
	return s.Content
}

type Rune rune

func (r *Rune) Accept(v Visitor) { v.VisitRune(r) }

func (r *Rune) UnmarshalJSON(data []byte) error {
	var r2 rune
	err := json.Unmarshal(data, &r2)
	if err != nil {
		return err
	}
	*r = Rune(r2)
	return nil
}

func (r *Rune) MarshalJSON() ([]byte, error) {
	return json.Marshal(rune(*r))
}

func (r *Rune) Info(_ *Env) util.Info                         { return util.InfoPure }
func (r *Rune) Eval(_ *Env, _ int) (result Evaler, err error) { return r, nil }
func (r *Rune) String() string                                { return string(*r) }
func (r *Rune) Inspect() string                               { return strconv.QuoteRune(rune(*r)) }
func (r *Rune) IDUses() []string                              { return nil }
func (r *Rune) IDSets() []string                              { return nil }
func (r *Rune) Capture(values []string) error {
	unquoted, err := strconv.Unquote(values[0])
	if err != nil {
		return err
	}
	if c := utf8.RuneCountInString(unquoted); c != 1 {
		return fmt.Errorf("%d runes in a rune literal", c)
	}
	for _, r2 := range unquoted {
		*r = Rune(r2)
		break
	}
	return nil
}
func (r *Rune) BecomeString() string { return string(*r) }
