package parser

import "gitlab.com/coalang/go-coa/try2/util"

func comp2(f func(a, b NumberLike) (bool, error)) Evaler {
	return NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
		ok, err := f(args[0].(BecomesNumberLike).BecomeNumberLike(), args[1].(BecomesNumberLike).BecomeNumberLike())
		if err != nil {
			return nil, err
		}
		return NewBool(ok), nil
	}, OptionArgs(TypeBecomesNumberLike, TypeBecomesNumberLike))
}

func comp(f func(a, b float64) bool) Evaler {
	return NewNative(util.InfoPure, func(env IEnv, args []Evaler) (Evaler, error) {
		return NewBool(f(float64(*args[0].(*Number)), float64(*args[1].(*Number)))), nil
	}, OptionArgs(TypeNumber, TypeNumber))
}
