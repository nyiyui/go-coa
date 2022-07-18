package parser

import (
	"gitlab.com/coalang/go-coa/try2/util"
	"os"
	"strings"
)

type SysEnv struct {
	_     util.NoCopy
	iterI int
}

func (o *SysEnv) Len() int { return len(o.Keys()) }

func (o *SysEnv) Index(i int) (key, value Evaler) {
	return NewString(o.Keys()[i]), NewString(o.Values()[i])
}

var _ Evaler = new(SysEnv)
var _ MapLike = new(SysEnv)
var _ Iter = new(SysEnv)

func (o *SysEnv) Info(_ *Env) util.Info                         { return util.InfoPure }
func (o *SysEnv) Eval(_ *Env, _ int) (result Evaler, err error) { return o, nil }
func (o *SysEnv) String() string                                { return "@sys_env" }
func (o *SysEnv) Inspect() string                               { return "@sys_env" }
func (o *SysEnv) IDUses() []string                              { return nil }
func (o *SysEnv) IDSets() []string                              { return nil }
func (o *SysEnv) Get(key string) (Evaler, bool, error) {
	value, ok := os.LookupEnv(key)
	return NewString(value), ok, nil
}

func (o *SysEnv) Set(key string, value Evaler) error { return os.Setenv(key, value.String()) }
func (o *SysEnv) Keys() (keys []string) {
	envvars := os.Environ()
	keys = make([]string, 0, len(envvars))
	for _, envvar := range envvars {
		if envvar != "" {
			keys = append(keys, strings.Split(envvar, "=")[0])
		}
	}
	return
}
func (o *SysEnv) Values() (values []string) {
	envvars := os.Environ()
	values = make([]string, 0, len(envvars))
	for _, envvar := range envvars {
		if envvar != "" {
			values = append(values, strings.Split(envvar, "=")[1])
		}
	}
	return
}
