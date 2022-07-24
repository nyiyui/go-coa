package parser

import (
	"encoding/json"
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Block struct {
	Pos     lexer.Position
	O       string `parser:"@OBrace"`
	Content Nodes  `parser:"@@"`
	C       string `parser:"@CBrace"`
}

func (b *Block) Accept(v Visitor) { v.VisitBlock(b) }

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

// TODO: expose uses that dont rely on $0, $1, etc
func (b *Block) IDUses() (re []string) {
	re = make([]string, 0)
	setByArguments := make([]string, 0)
	for i := 0; i < len(b.Content.Content); i++ {
		evaler := b.Content.Content[i].Select()
		sets := evaler.IDSets()
		uses := util.NoArguments(util.NoBuiltins(evaler.IDUses()))
		setByArguments = append(setByArguments, sets...)
		uses, _ = util.NoOverlap(uses, setByArguments)
		re = append(re, uses...)
	}
	return
}
func (b *Block) IDSets() []string { return nil /* only-outside-facing part matters*/ }

var _ Callable = new(Block)

func (b *Block) Info(env IEnv) util.Info     { return b.Content.Info(env) }
func (b *Block) Eval(_ IEnv) (Evaler, error) { return b, nil }

func (b *Block) Call(env IEnv, args []Evaler) (results Evaler, err error) {
	inner := env.Inherit(b.Pos)
	for i, arg := range args {
		inner.Def(fmt.Sprintf("$%d", i), arg)
	}
	b.Content.DisallowParallel = !b.runParallel()
	result, err := b.Content.Eval(inner)
	if err != nil {
		return ReturnVals(err)
	}
	return result, nil
}

func (b *Block) String() string       { return "{" + b.Content.String() + "}" }
func (b *Block) Inspect() string      { return "{" + b.Content.Inspect() + "}" }
func (b *Block) Pos_() lexer.Position { return b.Pos }
func (b *Block) runParallel() bool    { return len(b.O) != 2 }
