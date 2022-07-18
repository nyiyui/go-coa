package parser

import (
	"gitlab.com/coalang/go-coa/try2/errs"
)

func ReturnVals(err error) (Evaler, error) {
	if err, ok := err.(*errs.ERT); ok {
		evaler, err2 := ReturnVals(err.Err)
		if err2 == nil {
			return evaler, nil
		} else {
			return nil, err
		}
	}
	if err, ok := err.(*ErrReturn); ok {
		if err.Len >= 2 {
			err.Len--
			return nil, err
		}
		return err.Value, nil
	}
	return nil, err
}
