package oneofoneofdflt

type IsFoo interface { //want IsFoo:`oneofdflt\.A,oneofdflt\.B`
	isFoo()
}

type A struct{}

func (A) isFoo() {}

type B struct{}

func (B) isFoo() {}

// default branch should not be used for sum types as this
// breaks the exhaustiveness property
func oneofdflt(x IsFoo) {
	switch x.(type) { // want `missing default case on IsFoo: code cannot handle nil interface`
	case A, B:
	}
}
