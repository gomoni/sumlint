package test

type SumFoo interface {
	sumFoo()
}

type A struct{}

func (A) sumFoo() {}

type B struct{}

func (B) sumFoo() {}

func good(x SumFoo) {
	switch x.(type) {
	case A, B:
	default:
	}
}

func noDefault(x SumFoo) {
	switch x.(type) {
	case A, B:
	}
}

func noB(x SumFoo) {
	switch x.(type) {
	case A:
	default:
	}
}
