package test

import "github.com/gomoni/sumlint/test/sum"

func good(x sum.SumFoo) {
	switch x.(type) {
	case sum.A, sum.B:
	default:
	}
}

func noDefault(x sum.SumFoo) {
	switch x.(type) {
	case sum.A, sum.B:
	}
}

func noB(x sum.SumFoo) {
	switch x.(type) {
	case sum.A:
	default:
	}
}
