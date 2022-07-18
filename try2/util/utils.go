package util

import (
	"fmt"
)

func NoOverlap(a, b []string) ([]string, []string) {
	// convert to maps
	aM := map[string]struct{}{}
	bM := map[string]struct{}{}
	for _, v := range a {
		aM[v] = struct{}{}
	}
	for _, v := range b {
		bM[v] = struct{}{}
	}

	// remove overlaps
	var ok bool
	var k string
	for k = range aM {
		if _, ok = bM[k]; ok {
			delete(aM, k)
			delete(bM, k)
		}
	}

	// convert to slice
	c, d := make([]string, 0, len(aM)), make([]string, 0, len(bM))
	for k = range aM {
		c = append(c, k)
	}
	for k = range bM {
		d = append(d, k)
	}
	return c, d
}

func IsBuiltin(name string) bool  { return name[0] == '@' }
func IsArgument(name string) bool { return name[0] == '$' }

var (
	NoBuiltins  = makeFilter(IsBuiltin)
	NoArguments = makeFilter(IsArgument)

	allBuiltins  = makeAll(IsBuiltin)
	allArguments = makeAll(IsArgument)
)

func containsAll(a, b []string) bool {
	for _, bv := range b {
		if !contains(a, bv) {
			return false
		}
	}
	return true
}

func contains(a []string, b string) bool {
	for _, av := range a {
		if av == b {
			return true
		}
	}
	return false
}

func makeFilter(f func(string) bool) func([]string) []string {
	return func(names []string) []string {
		re := make([]string, 0, len(names))
		for _, name := range names {
			if !f(name) {
				re = append(re, name)
			}
		}
		return re
	}
}

func makeAll(f func(string) bool) func([]string) bool {
	return func(names []string) bool {
		for _, name := range names {
			if !f(name) {
				return false
			}
		}
		return true
	}
}

func EvalResources2(rs []ResourceDef, args []string) []Resource {
	re := make([]Resource, len(rs))
	for i, r := range rs {
		re[i].Name = r.Name
		if r.Arg != -1 {
			if r.Arg >= len(re) {
				re[i].Arg = "error"
			}
			re[i].Arg = args[r.Arg]
		}
	}
	return re
}

func EvalResources(rs []ResourceDef, args []string) ([]ResourceDef, []string) {
	re := make([]string, len(rs))
	for i, r := range rs {
		if r.Arg == -1 {
			re[i] = ""
		} else {
			if r.Arg >= len(re) {
				re[i] = "error"
			}
			re[i] = args[r.Arg]
		}
	}
	return rs, re
}

func ToInspect(stringer interface{ Inspect() string }) string {
	if stringer == nil {
		return ""
	}
	return stringer.Inspect()
}
func ToString(stringer fmt.Stringer) string {
	if stringer == nil {
		return ""
	}
	return stringer.String()
}

func StringSliceInterface(slice []interface{}) string {
	re := "("
	for i, thing := range slice {
		re += fmt.Sprintf("%T", thing)
		if i != len(slice)-1 {
			re += " "
		}
	}
	return re + ")"
}
