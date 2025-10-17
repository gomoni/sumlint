package sum

type SumFoo interface {
	sumFoo()
}

type A struct{}

func (A) sumFoo() {}

type B struct{}

func (B) sumFoo() {}
