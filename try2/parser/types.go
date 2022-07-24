package parser

import (
	"fmt"
	"time"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Visitor interface {
	VisitNodes(*Nodes)
	VisitNode(*Node)
	VisitNumber(*Number)
	VisitID(*ID)
	VisitString(*String)
	VisitRune(*Rune)
	VisitCall(*Call)
	VisitBlock(*Block)
	VisitList(*List)
}

type Thing interface {
	Accept(Visitor)
}

type maybeRunParallel interface {
	runParallel() bool
}

func isRunParallel(i interface{}) bool {
	r, ok := i.(maybeRunParallel)
	if !ok {
		return false
	}
	return r.runParallel()
}

type Evaler interface {
	HasInfo
	Eval(env IEnv) (result Evaler, err error)
	fmt.Stringer
	Inspect() string

	IDUses() []string
	IDSets() []string
}

type IEnv interface {
	Def(key string, evaler Evaler)
	Mod(key string, evaler Evaler)
	Get(key string) (Evaler, bool)
	Has(key string) bool
	HasKeys(keys []string) bool

	Dump() *EnvDump

	AddHook(name string, f hook)

	Keys() []string
	MyKeys() []string

	AllowParallel2() bool
	Debug2() bool
	Pos2() lexer.Position

	Printf(format string, v ...interface{})

	CheckResources(rs []util.Resource) bool
	BadResources(rs []util.Resource) []int
	LockResources(pos lexer.Position, rs []util.ResourceDef, args []string)
	UnlockResources(pos lexer.Position, rs []util.ResourceDef, args []string)

	Inherit(pos lexer.Position) IEnv
	InheritLone(pos lexer.Position) IEnv

	LoadPath(path string) (re Evaler, err error)
}

var _ IEnv = (*Env)(nil)

func Eval(e Evaler, env IEnv) (result Evaler, err error) {
	if e == nil {
		return nil, nil
	}
	//log.Println("eval",GetPos(e), e.IDUses())
	start := time.Now()
	switch e.(type) {
	case *Globber, *Regexer, *SysEnv, *Number, *Bool, *Rune, *Time:
	case *ID, *String:
	default:
		costCacheSet(e.Inspect(), uint64(time.Since(start).Nanoseconds()))
	}
	result, err = e.Eval(env)
	return
}

func GetPos(v interface{}) lexer.Position {
	if p, ok := v.(HasPos); ok {
		return p.Pos_()
	}
	return lexer.Position{}
}

type HasPos interface {
	Pos_() lexer.Position
}

type mayBeSpecial interface {
	isSpecial() bool
}

type Callable interface {
	Evaler
	HasInfo
	Call(env IEnv, args []Evaler) (results Evaler, err error)
}

type HasInfo interface {
	Info(env IEnv) util.Info
}

type Error struct {
	Pos    lexer.Position
	Nested error
}

func (e *Error) Unwrap() error {
	return e.Nested
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Nested)
}
