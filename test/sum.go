package test

import "github.com/gomoni/sumlint/test/sum"

type sumTest struct{}

func (sumTest) good(x sum.SumFoo) {
	switch x.(type) {
	case sum.A, *sum.B:
	default:
	}
}

func (sumTest) noDefault(x sum.SumFoo) {
	switch x.(type) {
	case sum.A, *sum.B:
	}
}

func (sumTest) noB(x sum.SumFoo) {
	switch x.(type) {
	case sum.A:
	default:
	}
}
