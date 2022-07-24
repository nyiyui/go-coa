package parser

import (
	"fmt"
	"regexp"
	"strconv"

	"gitlab.com/coalang/go-coa/try2/util"
)

type Regexer struct {
	_       util.NoCopy
	pattern *regexp.Regexp
	src     string
}

var _ Callable = new(Regexer)

func (r *Regexer) Info(_ IEnv) util.Info                  { return util.InfoPure }
func (r *Regexer) Eval(_ IEnv) (result Evaler, err error) { return r, nil }
func (r *Regexer) String() string                         { return fmt.Sprintf("(@regex %s)", strconv.Quote(r.src)) }
func (r *Regexer) Inspect() string                        { return r.String() }
func (r *Regexer) IDUses() []string                       { return nil }
func (r *Regexer) IDSets() []string                       { return nil }
func (r *Regexer) Call(env IEnv, args []Evaler) (results Evaler, err error) {
	args, err = OptionArgs(TypeBecomesString)(env, args)
	if err != nil {
		return nil, err
	}
	return NewBool(r.pattern.MatchString(args[0].(BecomesString).BecomeString())), nil
}
