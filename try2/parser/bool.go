package parser

import (
	"encoding/json"
	"fmt"

	"gitlab.com/coalang/go-coa/try2/util"
)

type Bool struct {
	Content bool
}

var _ Evaler = new(Bool)

func (b Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Content)
}

func NewBool(b bool) *Bool {
	return &Bool{Content: b}
}

func BoolFromEvaler(evaler Evaler) (bool, error) {
	switch evaler := evaler.(type) {
	case nil:
		return false, nil
	case *Bool:
		return evaler.Content, nil
	case *Number:
		return float64(*evaler) != 0, nil
	default:
		return false, fmt.Errorf("cannot use %s (type %T) as bool", evaler.Inspect(), evaler)
	}
}

var TypeBecomeBool = anyNilOf(TypeBool, TypeNumber)

func (b Bool) Info(_ *Env) util.Info                         { return util.InfoPure }
func (b Bool) Eval(_ *Env, _ int) (result Evaler, err error) { return b, nil }

func (b Bool) String() string {
	if b.Content {
		return "@true"
	}
	return "@false"
}

func (b Bool) Inspect() string  { return b.String() }
func (b Bool) IDUses() []string { return nil }
func (b Bool) IDSets() []string { return nil }
