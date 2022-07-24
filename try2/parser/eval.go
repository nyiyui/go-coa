package parser

import (
	"fmt"
	"strings"

	"gitlab.com/coalang/go-coa/try2/util"
)

type idProvider = func(*Call) []string

func idProviderNone(_ *Call) []string {
	return []string{}
}

var idSetsProviders = map[string]idProvider{}
var idUsesProviders = map[string]idProvider{}

func eval(env IEnv, allowParallel bool, evalers []Evaler) ([]Evaler, error) {
	err := checkEvalersRuntime(env, evalers)
	if err != nil {
		return nil, err
	}
	return eval_(env, allowParallel, evalers)
}

func checkEvalersRuntime(env IEnv, evalers []Evaler) (err error) {
	uses, sets := make([]string, 0), env.Keys()
	for _, evaler := range evalers {
		uses = append(uses, util.NoBuiltins(evaler.IDUses())...)
		sets = append(sets, util.NoBuiltins(evaler.IDSets())...)
		usesButDoesntExist, _ := util.NoOverlap(uses, sets)
		if len(usesButDoesntExist) > 0 {
			return fmt.Errorf("%s: required variables not defined: %s", GetPos(evaler), strings.Join(uses, ", "))
		}
	}
	var unusedSets []string
	for _, evaler := range evalers {
		uses, sets := util.NoOverlap(
			util.NoArguments(util.NoBuiltins(evaler.IDUses())),
			util.NoArguments(util.NoBuiltins(evaler.IDSets())),
		)
		_, unusedSets = util.NoOverlap(uses, append(unusedSets, sets...))
	}
	if len(unusedSets) > 0 {
		return fmt.Errorf("unused variables: %s", strings.Join(unusedSets, " "))
	}
	return
}

func eval_(env IEnv, _ bool, evalers []Evaler) ([]Evaler, error) {
	if len(evalers) == 0 {
		return []Evaler{}, nil
	}
	if env.AllowParallel2() &&
		len(evalers) != 1 && // no point in parallelizing a single function call
		evalersIsPure(env, evalers) {
		return evalParallel2(env, evalers)
	} else {
		return evalSeries(env, evalers)
	}
}

func evalSeries(env IEnv, evalers []Evaler) ([]Evaler, error) {
	if env.Debug2() {
		env.Printf("series %d", len(evalers))
	}
	var err error
	for i, evaler := range evalers {
		evalers[i], err = Eval(evaler, env)
		if err != nil {
			return nil, err
		}
	}
	return evalers, nil
}
