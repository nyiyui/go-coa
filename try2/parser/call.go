package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/errs"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Call struct {
	Pos     lexer.Position
	O       string `parser:"@OParen"`
	Content Nodes  `parser:"@@"`
	C       string `parser:"@CParen"`
}

func (c *Call) Accept(v Visitor) { v.VisitCall(c) }

func (c *Call) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

var _ Evaler = new(Call)

func (c *Call) Info(env IEnv) util.Info { return c.Content.Info(env) }
func (c *Call) ee(env IEnv) (Callable, error) {
	if len(c.Content.Content) == 0 {
		return nil, fmt.Errorf("blank call")
	}
	eeRaw := c.Content.Content[0].Select()
	eeRaw, err := Eval(eeRaw, env)
	if err != nil {
		return nil, err
	}
	{
		ee, ok := eeRaw.(Callable)
		if !ok {
			return nil, fmt.Errorf("%s not callable", eeRaw)
		}
		return ee, nil
	}
}
func (c *Call) Eval(env IEnv) (evaler Evaler, err error) {
	//env.printf("call init\t%s %s", c.Pos, c.Inspect())
	c2 := c
	defer func() {
		if err != nil {
			err = errs.AppendERT(err, errs.ERTFrame{
				Pos:  c.Pos,
				Call: util.ToInspect(c2),
			})
		}
	}()
	if len(c.Content.Content) == 0 {
		return nil, fmt.Errorf("%s: blank call", c.Pos)
	}
	ee, err := c.ee(env)
	if err != nil {
		return nil, err
	}
	isSpecial := false
	if ms, ok := ee.(mayBeSpecial); ok {
		isSpecial = ms.isSpecial()
	}

	args := c.Content.Select()[1:]
	if !isSpecial {
		args, err = eval(env, isRunParallel(ee), args)
		if err != nil {
			return nil, err
		}
	}

	c2 = &Call{
		Pos:     c.Pos,
		O:       c.O,
		Content: Nodes{Content: append([]Node{{Evaler: ee}}, toNodes(args)...)},
		C:       c.C,
	}

	if env.Debug2() {
		env.Printf("call args\t%s %s", c.Pos, c2.Inspect())
		env.Printf("%#v", ee.Info(env).Resources)
	}

	resources, stringsArgs := util.EvalResources(ee.Info(env).Resources, StringsSliceEvalers(args))
	rs := util.EvalResources2(ee.Info(env).Resources, StringsSliceEvalers(args))
	if !env.CheckResources(rs) {
		brs := env.BadResources(rs)
		brsString := make([]string, len(brs))
		for _, i := range brs {
			brsString = append(brsString, rs[i].String())
		}
		env.Printf("deferred evaluating %s due to usage of %d resource(s):%s", c.Pos, len(rs), strings.Join(brsString, " "))
		return c2, nil
	}
	env.LockResources(c.Pos, resources, stringsArgs)
	defer env.UnlockResources(c.Pos, resources, stringsArgs)
	result, err := ee.Call(env, args)
	if err != nil {
		return
	}
	if result == nil {
		return nil, errors.New("result of call is nil")
	}
	//env.printf("call re\t%s %s →  %s", c.Pos, c.Inspect(), result.Inspect())
	return result, nil
}
func (c *Call) String() string  { return "(" + c.Content.String() + ")" }
func (c *Call) Inspect() string { return "(" + c.Content.Inspect() + ")" }
func (c *Call) IDUses() []string {
	if fs := c.idUsesFor(); fs != nil {
		return fs
	}
	return c.Content.IDUses()
}
func (c *Call) IDSets() []string {
	if fs := c.idSetsFor(); fs != nil {
		return fs
	}
	return c.Content.IDSets()
}
func (c *Call) idUsesFor() []string {
	if len(c.Content.Content) == 0 {
		return nil
	}
	if id, ok := c.Content.Content[0].Select().(*ID); ok {
		if f, ok := idUsesProviders[id.Content]; ok {
			if f != nil {
				return f(c)
			}
		}
	}
	return nil
}
func (c *Call) idSetsFor() []string {
	if len(c.Content.Content) >= 2 {
		if id, ok := c.Content.Content[0].Select().(*ID); ok {
			if f, ok := idSetsProviders[id.Content]; ok {
				if f != nil {
					return f(c)
				}
			}
		}
	}
	return nil
}
func (c *Call) Pos_() lexer.Position { return c.Pos }
