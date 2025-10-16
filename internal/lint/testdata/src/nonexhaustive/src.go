package nonexhaustive

type SumFoo interface { //want SumFoo:`nonexhaustive\.A,nonexhaustive\.B,nonexhaustive.C`
	sumFoo()
}

type A struct{}

func (A) sumFoo() {}

type B struct{}

func (B) sumFoo() {}

type C struct{}

func (C) sumFoo() {}

// Non‑exhaustive: missing B, C.
func bad(x SumFoo) {
	switch x.(type) { // want `non-exhaustive type switch on SumFoo: missing cases for: nonexhaustive.B, nonexhaustive.C`
	case A:
	default:
	}
}
