package src

type Stack struct{}

type Element struct{}

func NewStack() *Stack { // want `the return value of the function has a different type than the declared type above.`
	return &Stack{}
}

func NewFoo() *Foo { // want `the constructor must be positioned after the type is defined`
	return &Foo{}
}

type Foo struct{}

func (f *Foo) Bar() int {
	return 42
}

func (s *Stack) Pop() {} // want `the method has a different type than the method above.`

type Queue []int

func (e Element) Get() {} // want `the method has a different type than the declared type above.`

type Tree int

func (t *Tree) Set() {} // want `the method must be located below the constructor function.`

func NewTree() *Tree {
	var t Tree
	return &t
}

func (e *Element) Set() {} // want `the method has a different type than the return value of the function above.`

type bTree string

func NewMis() *bTree {
	return nil
}

func smtFunc() {} // want `the function is located inside the constructor and method block.`

func (m *bTree) Get() {}

type set []map[int]int

func (s set) Get(i int) int {
	return 0
}

func smtWithSet() {} // want `the function is located inside a block of consecutive methods.`

func (s set) GetAll() {}
