package vm

import (
	"log"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/parser"
	"gitlab.com/coalang/go-coa/try2/util"
)

func (s *Scope) iEnv() parser.IEnv { return &iEnv{s: s} }

type iEnv struct {
	s *Scope
}

func (e *iEnv) Def(string, parser.Evaler)        { panic("unimplemented") }
func (e *iEnv) Mod(string, parser.Evaler)        { panic("unimplemented") }
func (e *iEnv) Get(string) (parser.Evaler, bool) { panic("unimplemented") }
func (e *iEnv) Has(string) bool                  { panic("unimplemented") }
func (e *iEnv) HasKeys([]string) bool            { panic("unimplemented") }

func (e *iEnv) Dump() *parser.EnvDump { panic("unimplemented") }

func (e *iEnv) AddHook(string, parser.Hook) { panic("unimplemented") }

func (e *iEnv) Keys() []string   { panic("unimplemented") }
func (e *iEnv) MyKeys() []string { panic("unimplemented") }

func (e *iEnv) AllowParallel2() bool { panic("unimplemented") }
func (e *iEnv) Debug2() bool         { panic("unimplemented") }
func (e *iEnv) Pos2() lexer.Position { return lexer.Position{Filename: e.s.pos} }

func (e *iEnv) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (e *iEnv) CheckResources(rs []util.Resource) bool { panic("unimplemented") }
func (e *iEnv) BadResources(rs []util.Resource) []int  { panic("unimplemented") }
func (e *iEnv) LockResources(pos lexer.Position, rs []util.ResourceDef, args []string) {
	panic("unimplemented")
}
func (e *iEnv) UnlockResources(pos lexer.Position, rs []util.ResourceDef, args []string) {
	panic("unimplemented")
}

func (e *iEnv) Inherit(pos lexer.Position) parser.IEnv     { panic("unimplemented") }
func (e *iEnv) InheritLone(pos lexer.Position) parser.IEnv { panic("unimplemented") }

func (e *iEnv) LoadPath(path string) (re parser.Evaler, err error) { panic("unimplemented") }
