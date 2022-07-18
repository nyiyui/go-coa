package parser

import (
	"encoding/json"
	"fmt"
	"gitlab.com/coalang/go-coa/try2/util"
)

type NativeFunc func(env *Env, args []Evaler) (Evaler, error)

type Native struct {
	_ util.NoCopy
	f       NativeFunc
	i       util.Info
	special bool
}

func (n *Native) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

var _ Callable = new(Native)
var _ mayBeSpecial = new(Native)

func nativeSpecial(name string, uses, sets idProvider, native *Native) *Native {
	idUsesProviders[name] = uses
	idSetsProviders[name] = sets
	native.special = true
	return native
}

func NewNative(info util.Info, native NativeFunc, options ...Option) *Native {
	return &Native{i: info, f: func(env *Env, args []Evaler) (Evaler, error) {
		var err error
		for _, option := range options {
			args, err = option(env, args)
			if err != nil {
				return nil, err
			}
		}
		return native(env, args)
	}}
}

func (n *Native) isSpecial() bool { return n.special }
func (n *Native) String() string {
	i := n.i.String()
	if i != "" {
		i = " " + i
	}
	return fmt.Sprintf("(@native %p%s)", n.f, i)
}
func (n *Native) Inspect() string                               { return n.String() }
func (n *Native) Info(_ *Env) util.Info                         { return n.i }
func (n *Native) Eval(_ *Env, _ int) (result Evaler, err error) { return n, nil }
func (n *Native) Call(env *Env, args []Evaler) (Evaler, error)  { return n.f(env, args) }
func (n *Native) IDUses() []string                              { return nil }
func (n *Native) IDSets() []string                              { return nil }
