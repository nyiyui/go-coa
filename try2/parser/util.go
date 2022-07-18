package parser

import (
	"fmt"
	"gitlab.com/coalang/go-coa/try2/util"
)

func StringsSliceEvalers(evalers []Evaler) []string {
	re := make([]string, len(evalers))
	for i, evaler := range evalers {
		re[i] = util.ToString(evaler)
	}
	return re
}

func StringSliceEvaler(slice []Evaler) string {
	re := "("
	for i, thing := range slice {
		re += fmt.Sprintf("%T", thing)
		if i != len(slice)-1 {
			re += " "
		}
	}
	return re + ")"
}
