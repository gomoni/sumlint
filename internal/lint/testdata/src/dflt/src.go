package dflt

type SumFoo interface { //want SumFoo:`dflt\.A,dflt\.B`
	sumFoo()
}

type A struct{}

func (A) sumFoo() {}

type B struct{}

func (B) sumFoo() {}

// default branch should not be used for sum types as this
// breaks the exhaustiveness property
func dflt(x SumFoo) {
	switch x.(type) { // want `missing default case on SumFoo: code cannot handle nil interface`
	case A, B:
	}
}
