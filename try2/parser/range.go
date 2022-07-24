package parser

import (
	"fmt"

	"gitlab.com/coalang/go-coa/try2/util"
)

type Range struct {
	start, stop, step int
}

func (r *Range) Info(_ IEnv) util.Info { return util.InfoPure }

func (r *Range) Eval(_ IEnv) (result Evaler, err error) { return r, nil }

func (r *Range) String() string { return fmt.Sprintf("(@range %d %d %d)", r.start, r.start, r.step) }

func (r *Range) Inspect() string { return r.String() }

func (r *Range) IDUses() []string { return nil }

func (r *Range) IDSets() []string { return nil }

var _ Iter = (*Range)(nil)

func (r *Range) Len() int {
	if r.step == 0 {
		r.step = 1
	}
	return (r.stop - r.start) / r.step
}

func (r *Range) Index(i int) (key, value Evaler) {
	return NewNumber(float64(i)), NewNumber(float64(r.start + r.step*i))
}
