package parser

import (
	"fmt"
	"gitlab.com/coalang/go-coa/try2/util"
	"math"
	"math/cmplx"
	"strconv"
)

type Complex complex128

var _ NumberLike = (*Complex)(nil)
var _ BecomesNumberLike = (*Complex)(nil)

func (c *Complex) BecomeNumberLike() NumberLike { return c }

func (c *Complex) Info(_ *Env) util.Info { return util.InfoPure }

func (c *Complex) Eval(_ *Env, _ int) (result Evaler, err error) { return c, nil }

func (c *Complex) Inspect() string { return c.String() }

func (c *Complex) IDUses() []string { return nil }

func (c *Complex) IDSets() []string { return nil }

func (c *Complex) Clone() NumberLike {
	c2 := *c
	return &c2
}

func (c *Complex) Add(n NumberLike) bool {
	switch n := n.(type) {
	case *Complex:
		*c += *n
	case *Int:
		*c += Complex(complex(float64(*n), 0))
	case *Float:
		*c += Complex(complex(float64(*n), 0))
	default:
		return false
	}
	return true
}

func (c *Complex) Sub(n NumberLike) bool {
	switch n := n.(type) {
	case *Complex:
		*c -= *n
	case *Int:
		*c -= Complex(complex(float64(*n), 0))
	case *Float:
		*c -= Complex(complex(float64(*n), 0))
	default:
		return false
	}
	return true
}

func (c *Complex) Mul(n NumberLike) bool {
	switch n := n.(type) {
	case *Complex:
		*c *= *n
	case *Int:
		*c *= Complex(complex(float64(*n), 0))
	case *Float:
		*c *= Complex(complex(float64(*n), 0))
	default:
		return false
	}
	return true
}

func (c *Complex) Div(n NumberLike) bool {
	switch n := n.(type) {
	case *Complex:
		*c /= *n
	case *Int:
		*c /= Complex(complex(float64(*n), 0))
	case *Float:
		*c /= Complex(complex(float64(*n), 0))
	default:
		return false
	}
	return true
}

func (c *Complex) Abs() bool {
	*c = Complex(complex(cmplx.Abs(complex128(*c)), 0))
	return true
}

func (c *Complex) Mod(_ NumberLike) bool {
	return false
}

func (c *Complex) Pow(n NumberLike) bool {
	switch n := n.(type) {
	case *Complex:
		*c = Complex(cmplx.Pow(complex128(*c), complex128(*n)))
	case *Int:
		*c = Complex(cmplx.Pow(complex128(*c), complex(float64(*n), 0)))
	case *Float:
		*c = Complex(cmplx.Pow(complex128(*c), complex(float64(*n), 0)))
	default:
		return false
	}
	return true
}

func (c *Complex) Cmp(n NumberLike) (int, bool) {
	switch n := n.(type) {
	case *Complex:
		if *c == *n {
			return 0, true
		} else {
			return 0, false
		}
	default:
		return 0, false
	}
}

func (c *Complex) String() string {
	return fmt.Sprintf(
		"(@complex %s %s ##%s##)",
		strconv.FormatFloat(real(complex128(*c)), 'f', -1, 64),
		strconv.FormatFloat(imag(complex128(*c)), 'f', -1, 64),
		strconv.FormatComplex(complex128(*c), 'f', -1, 64),
	)
}

type Float float64

var _ NumberLike = (*Float)(nil)
var _ BecomesNumberLike = (*Float)(nil)
var _ BecomesFloat64 = (*Float)(nil)

func (f *Float) BecomeFloat64() float64 { return float64(*f) }

func (f *Float) BecomeNumberLike() NumberLike { return f }

func (f *Float) Info(_ *Env) util.Info { return util.InfoPure }

func (f *Float) Eval(_ *Env, _ int) (result Evaler, err error) { return f, nil }

func (f *Float) Inspect() string { return f.String() }

func (f *Float) IDUses() []string { return nil }

func (f *Float) IDSets() []string { return nil }

func (f *Float) Clone() NumberLike {
	f2 := *f
	return &f2
}

func (f *Float) Add(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*f += *n
	case *Int:
		*f += Float(*n)
	default:
		return false
	}
	return true
}

func (f *Float) Sub(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*f -= *n
	case *Int:
		*f -= Float(*n)
	default:
		return false
	}
	return true
}

func (f *Float) Mul(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*f *= *n
	case *Int:
		*f *= Float(*n)
	default:
		return false
	}
	return true
}

func (f *Float) Div(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*f /= *n
	case *Int:
		*f /= Float(*n)
	default:
		return false
	}
	return true
}

func (f *Float) Abs() bool {
	*f = Float(math.Abs(float64(*f)))
	return true
}

func (f *Float) Mod(_ NumberLike) bool {
	return false
}

func (f *Float) Pow(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*f = Float(math.Pow(float64(*f), float64(*n)))
	case *Int:
		*f = Float(math.Pow(float64(*f), float64(*n)))
	default:
		return false
	}
	return true
}

func (f *Float) Cmp(n NumberLike) (int, bool) {
	switch n := n.(type) {
	case *Float:
		if *f > *n {
			return 1, true
		} else if *f < *n {
			return -1, true
		} else {
			return 0, true
		}
	case *Int:
		if *f > Float(*n) {
			return 1, true
		} else if *f < Float(*n) {
			return -1, true
		} else {
			return 0, true
		}
	default:
		return 0, false
	}
}

func (f *Float) String() string {
	return strconv.FormatFloat(float64(*f), 'f', -1, 64)
}

type Int int64

var _ NumberLike = (*Int)(nil)
var _ BecomesNumberLike = (*Int)(nil)
var _ BecomesFloat64 = (*Int)(nil)

func (i *Int) BecomeFloat64() float64 { return float64(*i) }

func (i *Int) BecomeNumberLike() NumberLike { return i }

func (i *Int) Info(_ *Env) util.Info { return util.InfoPure }

func (i *Int) Eval(_ *Env, _ int) (result Evaler, err error) { return i, nil }

func (i *Int) Inspect() string { return i.String() }

func (i *Int) IDUses() []string { return nil }

func (i *Int) IDSets() []string { return nil }

func (i *Int) Clone() NumberLike {
	i2 := *i
	return &i2
}

func (i *Int) Add(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*i += Int(*n)
	case *Int:
		*i += *n
	default:
		return false
	}
	return true
}

func (i *Int) Sub(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*i -= Int(*n)
	case *Int:
		*i -= *n
	default:
		return false
	}
	return true
}

func (i *Int) Mul(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*i *= Int(*n)
	case *Int:
		*i *= *n
	default:
		return false
	}
	return true
}

func (i *Int) Div(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*i /= Int(*n)
	case *Int:
		*i /= *n
	default:
		return false
	}
	return true
}

func (i *Int) Abs() bool {
	if *i < 0 {
		*i *= -1
	}
	return true
}

func (i *Int) Mod(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*i %= Int(*n)
	case *Int:
		*i %= *n
	default:
		return false
	}
	return true
}

func (i *Int) Pow(n NumberLike) bool {
	switch n := n.(type) {
	case *Float:
		*i = Int(math.Pow(float64(*i), float64(*n)))
	case *Int:
		*i = Int(math.Pow(float64(*i), float64(*n)))
	default:
		return false
	}
	return true
}

func (i *Int) Cmp(n NumberLike) (int, bool) {
	switch n := n.(type) {
	case *Float:
		if *i > Int(*n) {
			return 1, true
		} else if *i < Int(*n) {
			return -1, true
		} else {
			return 0, true
		}
	case *Int:
		if *i > *n {
			return 1, true
		} else if *i < *n {
			return -1, true
		} else {
			return 0, true
		}
	default:
		return 0, false
	}
}

func (i *Int) String() string {
	return strconv.FormatInt(int64(*i), 10)
}
