package parser

import (
	"fmt"
	"strconv"

	"github.com/gobwas/glob"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Globber struct {
	_       util.NoCopy
	pattern glob.Glob
	src     string
}

var _ Callable = new(Globber)

func (g *Globber) Info(_ IEnv) util.Info                  { return util.InfoPure }
func (g *Globber) Eval(_ IEnv) (result Evaler, err error) { return g, nil }
func (g *Globber) String() string                         { return fmt.Sprintf("(@glob %s)", strconv.Quote(g.src)) }
func (g *Globber) Inspect() string                        { return g.String() }
func (g *Globber) IDUses() []string                       { return nil }
func (g *Globber) IDSets() []string                       { return nil }
func (g *Globber) Call(env IEnv, args []Evaler) (results Evaler, err error) {
	args, err = OptionArgs(TypeBecomesString)(env, args)
	if err != nil {
		return nil, err
	}
	return NewBool(g.pattern.Match(args[0].(BecomesString).BecomeString())), nil
}
