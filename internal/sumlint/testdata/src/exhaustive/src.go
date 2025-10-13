package exhaustive

type SumFoo interface { //want SumFoo:`exhaustive\.A,exhaustive\.B`
	sumFoo()
}

type A struct{}

func (A) sumFoo() {}

type B struct{}

func (B) sumFoo() {}

// Exhaustive switch (A,B) – should produce no diagnostics.
func good(x SumFoo) {
	switch x.(type) {
	case A, B:
	default:
	}
}
