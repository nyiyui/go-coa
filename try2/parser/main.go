package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/participle/v2/lexer/stateful"
)

const allowed = `@$a-zA-Z_`
const ident = `[` + allowed + `][` + allowed + `0-9]*`

var lexerDef = lexer.Must(stateful.NewSimple([]stateful.Rule{
	{Name: "comment", Pattern: `#[^\n]*\n`}, // names starting with lowercase are elided
	{Name: "longComment", Pattern: `##[^#]##`},
	{Name: "space", Pattern: `[\s]+`},
	{Name: "ID", Pattern: ident},
	{Name: "String", Pattern: `\"(\\.|[^"\\])*\"`},
	{Name: "Rune", Pattern: `\'(\\.|[^'\\])*\'`},
	{Name: "Number", Pattern: `-?[0-9]+(\.[0-9]+)?`},
	{Name: "OBrack", Pattern: `\[m?`},
	{Name: "CBrack", Pattern: `\]`},
	{Name: "OBrace", Pattern: `\{%?`},
	{Name: "CBrace", Pattern: `\}`},
	{Name: "OParen", Pattern: `\(`},
	{Name: "CParen", Pattern: `\)`},
}))

var Parser = participle.MustBuild(&Nodes{}, participle.Lexer(lexerDef))
