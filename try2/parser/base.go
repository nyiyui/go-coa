package parser

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/gobwas/glob"
	"gitlab.com/coalang/go-coa/try2/util"
)

var ErrInternal = errors.New("internal error")
var ErrBreak = fmt.Errorf("break %w", ErrInternal)
var ErrContinue = fmt.Errorf("continue %w", ErrInternal)
var ErrUndefined = errors.New("operation undefined")

type ErrReturn struct {
	Len   int
	Value Evaler
}

func (e *ErrReturn) Error() string {
	return fmt.Sprintf("(@return_len %s %d)", util.ToInspect(e.Value), e.Len)
}

func (e *ErrReturn) Is(target error) bool {
	_, ok := target.(*ErrReturn)
	return ok
}

var OsArgs = os.Args

func NewBase() map[string]Evaler { return newBase() }

func newBase() map[string]Evaler {
	return map[string]Evaler{
		"@true":  &Bool{Content: true},
		"@false": &Bool{Content: false},

		"@include": nativeSpecial("@include", func(c *Call) []string {
			return c.Content.Content[1].Select().IDUses()
		}, func(c *Call) []string {
			return (&Nodes{Content: c.Content.Content[2:]}).IDUses()
		}, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (re Evaler, err error) {
			path := args[0].(BecomesString).BecomeString()
			defer func() {
				if err != nil {
					err = fmt.Errorf("including %s: %w", path, err)
				}
			}()
			inner := env.InheritLone(GetPos(args[0]))
			_, err = inner.LoadPath(filepath.Dir(env.Pos2().Filename) + "/" + path)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					libPath, ok := os.LookupEnv("COA_LIB")
					if !ok {
						libPath = "https://coalang.gitlab.io/lib/"
					}
					_, err = inner.LoadPath(libPath + path)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			}
			var value Evaler
			for _, key := range inner.MyKeys() {
				if util.IsBuiltin(key) {
					continue
				}
				value, _ = inner.Get(key)
				env.Def(key, value)
			}
			return nil, nil
		}, OptionArgsPrefix(TypeBecomesString))),

		"@time_now": NewNative(util.Info{Resources: []util.ResourceDef{{"os.time", -1}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			return &Time{Time: time.Now()}, nil
		}, OptionArgs()),
		"@time_sleep": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			time.Sleep(time.Duration(float64(*(args[0].(*Number))) * float64(time.Second)))
			return nil, nil
		}, OptionArgs(TypeNumber)),

		"@sys_os":   NewString(runtime.GOOS),
		"@sys_arch": NewString(runtime.GOARCH),
		"@sys_args": NewList(OsArgs),
		"@sys_exit": NewNative(util.Info{Resources: []util.ResourceDef{{"os.exit", -1}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			code := int(args[0].(*Number).BecomeFloat64())
			os.Exit(code)
			// panic there to replace return statement
			//goland:noinspection GoUnreachableCode
			panic(fmt.Sprintf("failed to exit with code %d", code))
		}, OptionArgs(TypeNumber)),
		"@sys_env": new(SysEnv),

		"@assert": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			ok, err := BoolFromEvaler(args[0])
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("failed assertion: %s", args[1].(BecomesString).BecomeString())
			}
			return args[0], nil
		}, OptionArgs(TypeAny, TypeBecomesString)),

		"@filter": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			nodes := args[0].(HasNodes).Nodes()
			filterer := args[1].(Callable)
			var evaler Evaler
			var err error
			var node Node
			var i int
			var keep bool
			defer func() {
				if err != nil {
					err = fmt.Errorf("@filter on index %d: %w", i, err)
				}
			}()
			re := make([]Node, 0)
			for i, node = range nodes.Content {
				evaler, err = filterer.Call(env, []Evaler{node.Select()})
				if err != nil {
					return nil, err
				}
				keep, err = BoolFromEvaler(evaler)
				if err != nil {
					return nil, err
				}
				if keep {
					re = append(re, nodes.Content[i])
				}
			}
			return &List{Content: Nodes{Content: re}}, nil
		}, OptionArgs(TypeHasNodes, TypeCallable)),
		"@glob": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			pattern := args[0].(BecomesString).BecomeString()
			compiled, err := glob.Compile(pattern)
			if err != nil {
				return nil, err
			}
			return &Globber{pattern: compiled, src: pattern}, nil
		}, OptionArgs(TypeBecomesString)),
		"@regex": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			pattern := args[0].(BecomesString).BecomeString()
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return nil, err
			}
			return &Regexer{pattern: compiled, src: pattern}, nil
		}, OptionArgs(TypeBecomesString)),

		"@error": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return nil, fmt.Errorf("error: %s", args[0].(*String).Content)
		}, OptionArgs(TypeString)),
		"@continue": nativeSpecial("@continue", idProviderNone, idProviderNone, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return nil, ErrContinue
		}, OptionNone)),
		"@break": nativeSpecial("@break", idProviderNone, idProviderNone, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return nil, ErrBreak
		}, OptionNone)),
		"@return": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return nil, &ErrReturn{Value: args[0]}
		}, OptionArgs(TypeAny)),
		"@return_len": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return nil, &ErrReturn{
				Len:   int(args[1].(BecomesFloat64).BecomeFloat64()),
				Value: args[0],
			}
		}, OptionArgs(TypeAny, TypeBecomesFloat64)),

		"@label": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return args[1], nil
		}, OptionArgs(TypeString, TypeAny)),

		"@use": NewNative(util.InfoPure, func(_ IEnv, _ []Evaler) (Evaler, error) { return NewBool(false), nil }),
		"@def": nativeSpecial("@def", func(c *Call) []string {
			return c.Content.Content[2].Select().IDUses()
		}, func(c *Call) []string {
			return []string{c.Content.Content[1].Select().(*ID).Content}
		}, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			if len(args) != 2 {
				return nil, errors.New("@mod: must have 2 args")
			}
			var err error
			args[1], err = Eval(args[1], env.Inherit(lexer.Position{Filename: "@def"}))
			if err != nil {
				return nil, err
			}
			name := (args[0].(*ID)).Content
			if util.IsBuiltin(name) {
				return nil, errors.New("cannot @def or @mod builtin names")
			}
			if util.IsArgument(name) {
				return nil, errors.New("cannot @def or @mod argument names")
			}
			env.Def(name, args[1])
			return args[1], nil
		}, OptionArgs(TypeID, TypeAny))),
		"@mod": nativeSpecial("@mod", func(c *Call) []string {
			return c.Content.Content[2].Select().IDUses()
		}, func(c *Call) []string {
			return []string{c.Content.Content[1].Select().(*ID).Content}
		}, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			if len(args) != 2 {
				return nil, errors.New("@mod: must have 2 args")
			}
			var err error
			args[1], err = Eval(args[1], env.Inherit(lexer.Position{Filename: "@def"}))
			if err != nil {
				return nil, err
			}
			name := args[0].(*ID).Content
			if !env.Has(name) {
				return nil, fmt.Errorf("cannot modify undefined variable %s", name)
			}
			if util.IsBuiltin(name) {
				return nil, errors.New("cannot @def or @mod builtin names")
			}
			if util.IsArgument(name) {
				return nil, errors.New("cannot @def or @mod argument names")
			}
			env.Mod(name, args[1])
			return args[1], nil
		}, OptionArgs(TypeID, TypeAny))),

		"@for": nativeSpecial("@for", nil, nil, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			init := args[0]
			condition := args[1]
			iter := args[2]
			callable := args[3].(Callable)
			inner := env.Inherit(GetPos(callable))
			_, err := Eval(init, inner)
			if err != nil {
				return nil, err
			}
			results := make([]Node, 0)
			var result Evaler
			for {
				eval, err := Eval(condition, inner)
				if err != nil {
					return nil, err
				}
				cont, err := BoolFromEvaler(eval)
				if err != nil {
					return nil, err
				}
				if !cont {
					break
				}
				contThis := false
				result, err = callable.Call(inner, []Evaler{})
				if err != nil {
					if errors.Is(err, ErrBreak) {
						break
					}
					if errors.Is(err, ErrContinue) {
						contThis = true
						goto iter
					}
					return nil, err
				}
			iter:
				results = append(results, Node{Evaler: result})
				_, err = Eval(iter, inner)
				if err != nil {
					if errors.Is(err, ErrBreak) {
						break
					}
					if errors.Is(err, ErrContinue) {
						continue
					}
					return nil, err
				}
				if contThis {
					continue
				}
			}
			return &List{Content: Nodes{Content: results}}, nil
		}, OptionArgs(TypeAny, TypeAny, TypeAny, TypeCallable))),
		"@while": nativeSpecial("@while", nil, nil, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			condition := args[0]
			callable := args[1].(Callable)
			inner := env.Inherit(GetPos(callable))
			results := make([]Node, 0)
			var result Evaler
			for {
				eval, err := Eval(condition, inner)
				if err != nil {
					return nil, err
				}
				cont, err := BoolFromEvaler(eval)
				if err != nil {
					return nil, err
				}
				if !cont {
					break
				}
				result, err = callable.Call(inner, []Evaler{})
				if err != nil {
					if errors.Is(err, ErrBreak) {
						break
					}
					if errors.Is(err, ErrContinue) {
						continue
					}
					return nil, err
				}
				results = append(results, Node{Evaler: result})
			}
			return &List{Content: Nodes{Content: results}}, nil
		}, OptionArgs(TypeAny, TypeCallable))),

		"@if": nativeSpecial("@if", nil, func(c *Call) []string {
			return c.Content.Content[0].Select().IDSets()
		}, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			if len(args) == 0 {
				return nil, errors.New("blank")
			}
			var evaler Evaler
			var err error
			var b bool
			for i := 0; i < len(args)-1; i += 2 {
				evaler, err = Eval(args[i], env)
				if err != nil {
					return nil, err
				}
				b, err = BoolFromEvaler(evaler)
				if err != nil {
					return nil, fmt.Errorf("%d: %w", i, err)
				}
				if b {
					return Eval(args[i+1], env)
				}
			}
			if len(args)%2 == 1 {
				return Eval(args[len(args)-1], env)
			}
			return NewNumber(0), nil
		}, OptionVariadic(TypeAny))),

		"@mapnokey": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			list := args[0].(HasNodes)
			callable := args[1].(Callable)
			results := make([]Node, len(list.Nodes().Content))
			var evaler Evaler
			var err error
			for i, node := range list.Nodes().Content {
				evaler, err = callable.Call(env, []Evaler{node.Select()})
				if err != nil {
					return nil, err
				}
				results[i] = Node{Evaler: evaler}
			}
			return &List{Content: Nodes{Content: results}}, nil
		}, OptionArgs(TypeHasNodes, TypeCallable)),
		"@map": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			list := args[0].(Iter)
			callable := args[1].(Callable)
			results := make([]Node, 0)
			var evaler Evaler
			var err error
			for i := 0; i < list.Len(); i++ {
				key, value := list.Index(i)
				evaler, err = callable.Call(env, []Evaler{key, value})
				if err != nil {
					return nil, err
				}
				results = append(results, Node{Evaler: evaler})
			}
			return &List{Content: Nodes{Content: results}}, nil
		}, OptionArgs(TypeIter, TypeCallable)),
		"@range": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			switch len(args) {
			case 1:
				stop := int(args[0].(BecomesFloat64).BecomeFloat64())
				return &Range{stop: stop}, nil
			case 2:
				start := int(args[0].(BecomesFloat64).BecomeFloat64())
				stop := int(args[1].(BecomesFloat64).BecomeFloat64())
				return &Range{start: start, stop: stop}, nil
			case 3:
				start := int(args[0].(BecomesFloat64).BecomeFloat64())
				stop := int(args[1].(BecomesFloat64).BecomeFloat64())
				step := int(args[2].(BecomesFloat64).BecomeFloat64())
				return &Range{start: start, stop: stop, step: step}, nil
			default:
				return nil, fmt.Errorf("1, 2, or 3 args required: %d provided", len(args))
			}
		}, OptionVariadic(TypeBecomesFloat64)),

		"@split": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			nodes := args[0].(HasNodes).Nodes()
			against := args[1]
			re := make([]Node, 0)
			prev := 0
			for i, node := range nodes.Content {
				if node.String() == against.String() {
					re = append(re, Node{List: &List{Content: Nodes{Content: nodes.Content[prev:i]}}})
					prev = i + 1
				}
			}
			re = append(re, Node{List: &List{Content: Nodes{Content: nodes.Content[prev:]}}})
			return &List{Content: Nodes{Content: re}}, nil
		}, OptionArgs(TypeHasNodes, TypeAny)),
		"@trim_prefix": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			against := args[0].(HasNodes).Nodes()
			prefix := args[1].(HasNodes).Nodes()
			if len(prefix.Content) > len(against.Content) {
				return args[0], nil
			}
			for i, value := range prefix.Content {
				if against.Content[i].Inspect() != value.Inspect() {
					return args[0], nil
				}
			}
			return &List{
				Content: Nodes{
					Content: against.Content[len(prefix.Content):],
				},
			}, nil
		}, OptionArgs(TypeHasNodes, TypeHasNodes)),
		"@has_prefix": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			against := args[0].(HasNodes).Nodes()
			prefix := args[1].(HasNodes).Nodes()
			if len(prefix.Content) > len(against.Content) {
				return NewBool(false), nil
			}
			for i, value := range prefix.Content {
				if against.Content[i].Inspect() != value.Inspect() {
					return NewBool(false), nil
				}
			}
			return NewBool(true), nil
		}, OptionArgs(TypeHasNodes, TypeHasNodes)),
		"@trim_suffix": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			against := args[0].(HasNodes).Nodes()
			suffix := args[1].(HasNodes).Nodes()
			if len(suffix.Content) > len(against.Content) {
				return args[0], nil
			}
			for i, value := range suffix.Content {
				if against.Content[len(against.Content)-i-1].Inspect() != value.Inspect() {
					return args[0], nil
				}
			}
			return &List{
				Content: Nodes{
					Content: against.Content[:len(against.Content)-len(suffix.Content)],
				},
			}, nil
		}, OptionArgs(TypeHasNodes, TypeHasNodes)),
		"@has_suffix": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			against := args[0].(HasNodes).Nodes()
			suffix := args[1].(HasNodes).Nodes()
			if len(suffix.Content) > len(against.Content) {
				return NewBool(false), nil
			}
			for i, value := range suffix.Content {
				if against.Content[len(against.Content)-i-1].Inspect() != value.Inspect() {
					return NewBool(false), nil
				}
			}
			return NewBool(true), nil
		}, OptionArgs(TypeHasNodes, TypeHasNodes)),
		"@len": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return NewNumber(float64(len(args[0].(HasNodes).Nodes().Content))), nil
		}, OptionArgs(TypeHasNodes)),

		"@foldl": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			callable := args[0].(Callable)
			list := args[1].(HasNodes)
			var result Evaler
			var err error
			for i, node := range list.Nodes().Content {
				if i == 0 {
					result = node.Select()
					continue
				}
				result, err = callable.Call(env, []Evaler{node.Select(), result})
				if err != nil {
					return nil, err
				}
			}
			return result, nil
		}, OptionArgs(TypeCallable, TypeHasNodes)),
		"@foldr": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			callable := args[0].(Callable)
			list := args[1].(HasNodes)
			var result Evaler
			var err error
			for i, node := range list.Nodes().Content {
				if i == 0 {
					result = node.Select()
					continue
				}
				result, err = callable.Call(env, []Evaler{result, node.Select()})
				if err != nil {
					return nil, err
				}
			}
			return result, nil
		}, OptionArgs(TypeCallable, TypeHasNodes)),

		"@lt": comp2(func(a, b NumberLike) (bool, error) {
			cmp, ok := a.Cmp(b)
			if !ok {
				return false, ErrUndefined
			}
			return cmp == -1, nil
		}),
		"@le": comp2(func(a, b NumberLike) (bool, error) {
			cmp, ok := a.Cmp(b)
			if !ok {
				return false, ErrUndefined
			}
			return cmp == -1 || cmp == 0, nil
		}),
		"@gt": comp2(func(a, b NumberLike) (bool, error) {
			cmp, ok := a.Cmp(b)
			if !ok {
				return false, ErrUndefined
			}
			return cmp == +1, nil
		}),
		"@ge": comp2(func(a, b NumberLike) (bool, error) {
			cmp, ok := a.Cmp(b)
			if !ok {
				return false, ErrUndefined
			}
			return cmp == +1 || cmp == 0, nil
		}),
		"@eq": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return NewBool(args[0].Inspect() == args[1].Inspect()), nil
		}, OptionArgs(TypeAny, TypeAny)),
		"@ne": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return NewBool(args[0].Inspect() != args[1].Inspect()), nil
		}, OptionArgs(TypeAny, TypeAny)),
		"@or": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a, err := BoolFromEvaler(args[0])
			if err != nil {
				return nil, err
			}
			b, err := BoolFromEvaler(args[1])
			if err != nil {
				return nil, err
			}
			return NewBool(a || b), nil
		}, OptionArgs(TypeBecomeBool, TypeBecomeBool)),
		"@and": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a, err := BoolFromEvaler(args[0])
			if err != nil {
				return nil, err
			}
			b, err := BoolFromEvaler(args[1])
			if err != nil {
				return nil, err
			}
			return NewBool(a && b), nil
		}, OptionArgs(TypeBecomeBool, TypeBecomeBool)),
		"@not": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			b, err := BoolFromEvaler(args[0])
			if err != nil {
				return nil, err
			}
			return NewBool(b), nil
		}, OptionArgs(TypeBecomeBool)),

		"@concat": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			switch arg0 := args[0].(type) {
			case BecomesFloat64:
				switch arg1 := args[1].(type) {
				case BecomesFloat64:
					return NewNumber(arg0.BecomeFloat64() + arg1.BecomeFloat64()), nil
				case BecomesString:
					return nil, errors.New("must have same type for args")
				case nil:
					return args[0], nil
				}
			case BecomesString:
				switch arg1 := args[1].(type) {
				case BecomesFloat64:
					return nil, errors.New("must have same type for args")
				case BecomesString:
					return NewString(arg0.(BecomesString).BecomeString() + arg1.(BecomesString).BecomeString()), nil
				case nil:
					return args[0], nil
				}
			case nil:
				return args[1], nil
			}
			panic(fmt.Sprintf("unexpected type %T and %T", args[0], args[1]))
		}, OptionArgs(anyNilOf(TypeNumber, TypeBecomesString), anyNilOf(TypeNumber, TypeBecomesString))),

		"@add": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesNumberLike).BecomeNumberLike().Clone()
			b := args[1].(BecomesNumberLike).BecomeNumberLike()
			ok := a.Add(b)
			if !ok {
				return nil, ErrUndefined
			}
			return a, nil
		}, OptionArgs(TypeBecomesNumberLike, TypeBecomesNumberLike)),
		"@sub": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesNumberLike).BecomeNumberLike().Clone()
			b := args[1].(BecomesNumberLike).BecomeNumberLike()
			ok := a.Sub(b)
			if !ok {
				return nil, ErrUndefined
			}
			return a, nil
		}, OptionArgs(TypeBecomesNumberLike, TypeBecomesNumberLike)),
		"@mul": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesNumberLike).BecomeNumberLike().Clone()
			b := args[1].(BecomesNumberLike).BecomeNumberLike()
			ok := a.Mul(b)
			if !ok {
				return nil, ErrUndefined
			}
			return a, nil
		}, OptionArgs(TypeBecomesNumberLike, TypeBecomesNumberLike)),
		"@div": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesNumberLike).BecomeNumberLike().Clone()
			b := args[1].(BecomesNumberLike).BecomeNumberLike()
			ok := a.Div(b)
			if !ok {
				return nil, ErrUndefined
			}
			return a, nil
		}, OptionArgs(TypeBecomesNumberLike, TypeBecomesNumberLike)),
		"@rem": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesNumberLike).BecomeNumberLike().Clone()
			b := args[1].(BecomesNumberLike).BecomeNumberLike()
			ok := a.Mod(b)
			if !ok {
				return nil, ErrUndefined
			}
			return a, nil
		}, OptionArgs(TypeBecomesNumberLike, TypeBecomesNumberLike)),
		"@pow": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesNumberLike).BecomeNumberLike().Clone()
			b := args[1].(BecomesNumberLike).BecomeNumberLike()
			ok := a.Pow(b)
			if !ok {
				return nil, ErrUndefined
			}
			return a, nil
		}, OptionArgs(TypeBecomesNumberLike, TypeBecomesNumberLike)),
		"@abs": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesNumberLike).BecomeNumberLike().Clone()
			ok := a.Abs()
			if !ok {
				return nil, ErrUndefined
			}
			if a, ok := a.(*Complex); ok {
				f := Float(real(complex128(*a)))
				return &f, nil
			}
			return a, nil
		}, OptionArgs(TypeBecomesNumberLike)),

		"@http_get": NewNative(util.Info{Resources: []util.ResourceDef{{"http", 0}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			got, err := (&http.Client{Timeout: 10 * time.Second}).Get(args[0].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			body, err := io.ReadAll(got.Body)
			if err != nil {
				return nil, err
			}
			return NewString(string(body)), nil
		}, OptionArgs(TypeBecomesString)),

		"@file_write": NewNative(util.Info{Resources: []util.ResourceDef{{"fs.local", 0}}}, func(env IEnv, args []Evaler) (evaler Evaler, err error) {
			file, err := os.Create(args[0].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			defer func(file *os.File) {
				err2 := file.Close()
				if err2 != nil {
					err = err2
				}
			}(file)
			writeString, err := file.WriteString(args[1].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			return NewNumber(float64(writeString)), nil
		}, OptionArgs(TypeBecomesString, TypeBecomesString)),
		"@file_read": NewNative(util.Info{Resources: []util.ResourceDef{{"fs.local", 0}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			file, err := os.ReadFile(args[0].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			return NewString(string(file)), nil
		}, OptionArgs(TypeBecomesString)),
		"@file_remove": NewNative(util.Info{Resources: []util.ResourceDef{{"fs.local", 0}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			err := os.RemoveAll(args[0].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			return nil, nil
		}, OptionArgs(TypeBecomesString)),
		"@file_list": NewNative(util.Info{Resources: []util.ResourceDef{{"fs.local", 0}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			p := args[0].(BecomesString).BecomeString()
			dirs, err := os.ReadDir(p)
			if err != nil {
				return nil, err
			}
			re := &List{Content: Nodes{Content: make([]Node, len(dirs))}}
			for i, dir := range dirs {
				re.Content.Content[i] = toNode(NewString(dir.Name()))
			}
			return re, nil
		}, OptionArgs(TypeBecomesString)),

		"@io_out": NewNative(util.Info{Resources: []util.ResourceDef{{"io.stdout", -1}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			wrote, err := os.Stdout.WriteString(args[0].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			return NewNumber(float64(wrote)), nil
		}, OptionArgs(TypeBecomesString)),
		"@io_outln": NewNative(util.Info{Resources: []util.ResourceDef{{"io.stdout", -1}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			wrote, err := os.Stdout.WriteString(args[0].(BecomesString).BecomeString() + "\n")
			if err != nil {
				return nil, err
			}
			return NewNumber(float64(wrote)), nil
		}, OptionArgs(TypeBecomesString)),
		"@io_in": NewNative(util.Info{[]util.ResourceDef{{"io.stdin", -1}}}, func(env IEnv, args []Evaler) (Evaler, error) {
			reader := bufio.NewReader(os.Stdin)
			read, err := reader.ReadString(byte(*(args[0].(*Rune))))
			if err != nil {
				return nil, err
			}
			return NewString(read[:len(read)-1]), nil
		}, OptionArgs(TypeRune)),

		"@complex": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			a := args[0].(BecomesFloat64).BecomeFloat64()
			b := args[1].(BecomesFloat64).BecomeFloat64()
			c := Complex(complex(a, b))
			return &c, nil
		}, OptionArgs(TypeBecomesFloat64, TypeBecomesFloat64)),
		"@complex_from": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			c, err := strconv.ParseComplex(strings.TrimSpace(args[0].(BecomesString).BecomeString()), 64)
			if err != nil {
				return nil, err
			}
			{
				c := Complex(c)
				return &c, nil
			}
		}, OptionArgs(TypeBecomesString)),
		"@int": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			float, err := strconv.ParseInt(strings.TrimSpace(args[0].(BecomesString).BecomeString()), 10, 64)
			if err != nil {
				return nil, err
			}
			return NewNumber(float64(float)), nil
		}, OptionArgs(TypeBecomesString)),
		"@uint": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			float, err := strconv.ParseInt(strings.TrimSpace(args[0].(BecomesString).BecomeString()), 10, 64)
			if err != nil {
				return nil, err
			}
			return NewNumber(float64(float)), nil
		}, OptionArgs(TypeBecomesString)),
		"@float": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			float, err := strconv.ParseFloat(strings.TrimSpace(args[0].(BecomesString).BecomeString()), 64)
			if err != nil {
				return nil, err
			}
			return NewNumber(float), nil
		}, OptionArgs(TypeBecomesString)),
		"@string": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			if args[0] == nil {
				return NewString("<nil>"), nil
			}
			return NewString(args[0].String()), nil
		}, OptionArgs(TypeAny)),
		"@inspect": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			if args[0] == nil {
				return NewString("<nil>"), nil
			}
			return NewString(args[0].Inspect()), nil
		}, OptionArgs(TypeAny)),

		"@json_to": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			marshalled, err := json.Marshal(args[0])
			if err != nil {
				return nil, fmt.Errorf("json: %s", err)
			}
			return NewString(string(marshalled)), nil
		}, OptionArgs(TypeAny)),
		"@json_from": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			var v interface{}
			err := json.Unmarshal([]byte(args[0].(*String).Content), &v)
			if err != nil {
				return nil, fmt.Errorf("json: %s", err)
			}
			return toEvaler(v)
		}, OptionArgs(TypeString)),

		"@url_from_path": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			unescaped, err := url.PathUnescape(args[0].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			return NewString(unescaped), nil
		}, OptionArgs(TypeBecomesString)),
		"@url_from_query": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			unescaped, err := url.QueryUnescape(args[0].(BecomesString).BecomeString())
			if err != nil {
				return nil, err
			}
			return NewString(unescaped), nil
		}, OptionArgs(TypeBecomesString)),
		"@url_to_path": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return NewString(url.PathEscape(args[0].(BecomesString).BecomeString())), nil
		}, OptionArgs(TypeBecomesString)),
		"@url_to_query": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			return NewString(url.QueryEscape(args[0].(BecomesString).BecomeString())), nil
		}, OptionArgs(TypeBecomesString)),

		"@get_try": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			m := args[0].(MapLike)
			key := args[1].(*String).Content
			fallback := args[2]
			value, ok, err := m.Get(key)
			if err != nil {
				return nil, err
			}
			if !ok {
				return fallback, nil
			}
			return value, nil
		}, OptionArgs(TypeMapLike, TypeString, TypeAny)),
		"@get": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			m := args[0].(MapLike)
			key := args[1].(BecomesString).BecomeString()
			value, ok, err := m.Get(key)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("value for key %s not found", strconv.Quote(key))
			}
			return value, nil
		}, OptionArgs(TypeMapLike, TypeBecomesString)),
		"@set": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			m := args[0].(MapLike)
			key := args[1].(BecomesString).BecomeString()
			value := args[2]
			err := m.Set(key, value)
			if err != nil {
				return nil, err
			}
			return value, nil
		}, OptionArgs(TypeMapLike, TypeBecomesString, TypeAny)),
		"@keys": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			m := args[0].(MapLike)
			keys := m.Keys()
			l := &List{Content: Nodes{Content: make([]Node, 0, len(keys))}}
			for _, key := range keys {
				l.Content.Content = append(l.Content.Content, Node{Evaler: NewString(key)})
			}
			return l, nil
		}, OptionArgs(TypeMapLike)),

		"@select": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			list := args[0].(HasNodes)
			index := int(*(args[1].(*Number)))
			if l := len(list.Nodes().Content); index >= l {
				return nil, fmt.Errorf("index out of range: (@select [length of %d] %d)", l, index)
			} else if index < 0 {
				index = l + index
			}
			return list.Nodes().Content[index].Select(), nil
		}, OptionArgs(TypeHasNodes, TypeNumber)),
		"@take_from": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			list := args[0].(HasNodes)
			start := int(*(args[1].(*Number)))
			nodes := list.Nodes()
			if start < 0 {
				start = len(nodes.Content) + start
			}
			if l := len(nodes.Content); start >= l {
				return nil, fmt.Errorf("start out of range: %d of %d", start, l)
			}
			return &List{Content: Nodes{Pos: GetPos(list), Content: nodes.Content[start:]}}, nil
		}, OptionArgs(TypeHasNodes, TypeNumber)),
		"@take_to": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			list := args[0].(HasNodes)
			end := int(*(args[1].(*Number)))
			nodes := list.Nodes()
			if end < 0 {
				end = len(nodes.Content) + end
			}
			if l := len(nodes.Content); end > l {
				return nil, fmt.Errorf("end out of range: %d of %d", end, l)
			}
			return &List{Content: Nodes{Pos: GetPos(list), Content: nodes.Content[:end]}}, nil
		}, OptionArgs(TypeHasNodes, TypeNumber)),
		"@take": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			list := args[0].(HasNodes)
			start := int(*(args[1].(*Number)))
			end := int(*(args[2].(*Number)))
			nodes := list.Nodes()
			if start < 0 {
				start = len(nodes.Content) + start
			}
			if end < 0 {
				end = len(nodes.Content) + end
			}
			if start >= end {
				return nil, fmt.Errorf("start (%d) must be lesser than end (%d); start is equal to or larger than end", start, end)
			}
			if l := len(nodes.Content); start >= l {
				return nil, fmt.Errorf("start out of range: %d of %d", start, l)
			}
			if l := len(nodes.Content); end > l {
				return nil, fmt.Errorf("end out of range: %d of %d", end, l)
			}
			return &List{Content: Nodes{Pos: GetPos(list), Content: nodes.Content[start:end]}}, nil
		}, OptionArgs(TypeHasNodes, TypeNumber, TypeNumber)),

		"@bifrost_peek": nativeSpecial("@bifrost_peek", idProviderNone, idProviderNone, NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			Bifrost.Peek(env, args)
			return nil, nil
		}, OptionVariadic(TypeAny))),
		"@bifrost_profile": NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
			Bifrost.Profile(env)
			return nil, nil
		}, OptionArgs()),
	}
}
