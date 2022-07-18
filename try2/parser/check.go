package parser

import (
	"fmt"
	"gitlab.com/coalang/go-coa/try2/util"
	"strings"
)

func check(nodes *Nodes) error {
	err := checkEvalers(nodes.Select())
	if err != nil {
		return err
	}
	return nil
}

func checkEvalers(evalers []Evaler) (err error) {
	uses, sets := []string{}, []string{}
	for i, evaler := range evalers {
		uses = append(uses, evaler.IDUses()...)
		sets = append(sets, evaler.IDSets()...)
		uses, sets = util.NoBuiltins(uses), util.NoBuiltins(sets)
		uses, _ = util.NoOverlap(util.NoBuiltins(uses), util.NoBuiltins(sets))
		if len(uses) > 0 {
			return fmt.Errorf("%s: index %d: required variables not defined: %s", GetPos(evaler), i, strings.Join(uses, " "))
		}
	}
	return nil
}
