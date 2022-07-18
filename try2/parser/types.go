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
	Eval(env *Env, order int) (result Evaler, err error)
	fmt.Stringer
	Inspect() string

	IDUses() []string
	IDSets() []string
}

func Eval(e Evaler, env *Env) (result Evaler, err error) {
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
	result, err = e.Eval(env, 0)
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
	Call(env *Env, args []Evaler) (results Evaler, err error)
}

type HasInfo interface {
	Info(env *Env) util.Info
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
