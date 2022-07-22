package vm

import (
	"errors"

	"gitlab.com/coalang/go-coa/try2/compile"
)

type Program struct {
	insts []compile.Instruction
}

type Value interface{}

type VM struct {
}

func (v *VM) newScope() *Scope {
	return &Scope{
		vars: make([]Value, 0),
	}
}

func (v *VM) Execute(prog Program) (result Value, err error) {
	return v.newScope().exec(prog)
}

type Scope struct {
	vars []Value
}

func (s *Scope) exec(p Program) (res Value, err error) {
	return nil, errors.New("not implemented")
}
