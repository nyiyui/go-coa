package parser

import (
	"fmt"
	"reflect"

	"gitlab.com/coalang/go-coa/try2/util"
)

type Option func(env IEnv, args []Evaler) ([]Evaler, error)

var (
	TypeID                = new(ID)
	TypeBool              = new(Bool)
	TypeNumber            = new(Number)
	TypeNumberLike        = special(func(evaler Evaler) bool { _, ok := evaler.(NumberLike); return ok })
	TypeBecomesNumberLike = special(func(evaler Evaler) bool { _, ok := evaler.(BecomesNumberLike); return ok })
	TypeString            = new(String)
	TypeBecomesString     = special(func(evaler Evaler) bool { _, ok := evaler.(BecomesString); return ok })
	TypeBecomesFloat64    = special(func(evaler Evaler) bool { _, ok := evaler.(BecomesFloat64); return ok })
	TypeRune              = new(Rune)
	TypeCallable          = special(func(evaler Evaler) bool { _, ok := evaler.(Callable); return ok })
	TypeAny               = special(func(Evaler) bool { return true })
	TypeMap               = new(Map)
	TypeHasNodes          = special(func(evaler Evaler) bool { _, ok := evaler.(HasNodes); return ok })
	TypeIter              = special(func(evaler Evaler) bool { _, ok := evaler.(Iter); return ok })
	TypeMapLike           = special(func(evaler Evaler) bool { _, ok := evaler.(MapLike); return ok })
)

type NumberLike interface {
	Evaler
	Clone() NumberLike
	Add(NumberLike) bool
	Sub(NumberLike) bool
	Mul(NumberLike) bool
	Div(NumberLike) bool
	Mod(NumberLike) bool
	Pow(NumberLike) bool
	Abs() bool
	Cmp(NumberLike) (int, bool) // if a < b: -1; if a == b: 0; if a > b: +1
	fmt.Stringer
}

type BecomesNumberLike interface {
	BecomeNumberLike() NumberLike
}

type BecomesString interface {
	BecomeString() string
}

type BecomesFloat64 interface {
	BecomeFloat64() float64
}

type HasNodes interface {
	Nodes() Nodes
}

type Iter interface {
	Len() int
	Index(i int) (key, value Evaler)
}

func anyNilOf(ts ...interface{}) special {
	return func(evaler Evaler) bool {
		if evaler == nil {
			return true
		}
		for _, t := range ts {
			var argOk bool
			if at, ok := t.(special); ok {
				argOk = at(evaler)
			} else {
				argOk = util.ToString(reflect.TypeOf(t)) == util.ToString(reflect.TypeOf(evaler))
			}
			if argOk {
				return true
			}
		}
		return false
	}
}
func anyOf(ts ...interface{}) special {
	return func(evaler Evaler) bool {
		if evaler == nil {
			return false
		}
		for _, t := range ts {
			var argOk bool
			if at, ok := t.(special); ok {
				argOk = at(evaler)
			} else {
				argOk = util.ToString(reflect.TypeOf(t)) == util.ToString(reflect.TypeOf(evaler))
			}
			if argOk {
				return true
			}
		}
		return false
	}
}

type special func(Evaler) bool

func OptionArgsPrefix(argTypes ...interface{}) Option {
	return func(env IEnv, args []Evaler) (_ []Evaler, err error) {
		if len(argTypes) > len(args) {
			return nil, fmt.Errorf("wanted %s+, got %s (length)", util.StringSliceInterface(argTypes), StringSliceEvaler(args))
		}
		var arg Evaler
		for i, argType := range argTypes {
			arg = args[i]
			var argOk bool
			if at, ok := argType.(special); ok {
				argOk = at(arg)
			} else {
				argOk = util.ToString(reflect.TypeOf(argType)) == util.ToString(reflect.TypeOf(arg))
			}
			if !argOk {
				return nil, fmt.Errorf("wanted %s, got %s (%d)", util.StringSliceInterface(argTypes), StringSliceEvaler(args), i)
			}
		}
		return args, nil
	}
}

func OptionNone(_ IEnv, args []Evaler) ([]Evaler, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("wanted 0, got %s (length)", StringSliceEvaler(args))
	}
	return nil, nil
}

func OptionArgs(argTypes ...interface{}) Option {
	return func(env IEnv, args []Evaler) (_ []Evaler, err error) {
		if len(argTypes) != len(args) {
			return nil, fmt.Errorf("wanted %s, got %s (length)", util.StringSliceInterface(argTypes), StringSliceEvaler(args))
		}
		var arg Evaler
		for i, argType := range argTypes {
			arg = args[i]
			var argOk bool
			if at, ok := argType.(special); ok {
				argOk = at(arg)
			} else {
				argOk = util.ToString(reflect.TypeOf(argType)) == util.ToString(reflect.TypeOf(arg))
			}
			if !argOk {
				return nil, fmt.Errorf("wanted %s, got %s (%d)", util.StringSliceInterface(argTypes), StringSliceEvaler(args), i)
			}
		}
		return args, nil
	}
}

func OptionVariadic(argType interface{}) Option {
	return func(env IEnv, args []Evaler) ([]Evaler, error) {
		for i, arg := range args {
			var argOk bool
			if at, ok := argType.(special); ok {
				argOk = at(arg)
			} else {
				argOk = reflect.TypeOf(argType).String() == reflect.TypeOf(arg).String()
			}
			if !argOk {
				return nil, fmt.Errorf("wanted variadic %s, got %s (%d)", argType, StringSliceEvaler(args), i)
			}
		}
		return args, nil
	}
}
