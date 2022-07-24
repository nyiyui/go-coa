package parser

import (
	"time"

	"gitlab.com/coalang/go-coa/try2/util"
)

type Time struct {
	util.NoCopy
	time.Time
}

func (t *Time) Info(_ IEnv) util.Info                  { return util.InfoPure }
func (t *Time) Eval(_ IEnv) (result Evaler, err error) { return t, nil }
func (t *Time) String() string                         { return t.Time.String() }
func (t *Time) Inspect() string                        { return t.Time.String() }
func (t *Time) IDUses() []string                       { return nil }
func (t *Time) IDSets() []string                       { return nil }
func (t *Time) BecomeString() string                   { return t.Time.String() }
func (t *Time) BecomeFloat64() float64                 { return float64(t.Time.Unix()) }
