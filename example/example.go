package main

import (
	"github.com/githubnemo/pdump"
)

func Test1(in int, b []byte, in2 int, m map[string]int) {
	PrintInputs(Test1)
}

func Test2(in float64, s string) {
	PrintInputs(Test2)
}

func Test3(in int) (int, int) {
	PrintInputs(Test3)
	defer PrintOutputs(Test3)
	return 3, 4
}

func Test4(a [3]byte, t testStruct) {
	PrintInputs(Test4)
}

func Test5(t bool, c1 chan byte, c2 <-chan byte) {
	PrintInputs(Test5)
}

func Test6(c1 complex64, c2 complex128) {
	PrintInputs(Test6)
}

func Test7(f func(i int) bool) {
	PrintInputs(Test7)
}

func Test8(a interface{}, p *[]byte, up unsafe.Pointer) {
	PrintInputs(Test8)
}

type testStruct struct {
	t int
	b []byte
}

func main() {
	b := []byte{'A', 'B', 'C'}
	m := map[string]int{"foo": 3, "bar": 7}
	m["bar"] += 2
	s := "AAAA"
	a := [3]byte{'a', 'b', 'c'}
	c1 := make(chan byte)
	c2 := make(<-chan byte)
	p := &b
	up := unsafe.Pointer(p)
	Test1(2, b, 9, m)
	Test2(3.14, s)
	Test3(42)
	Test4(a, testStruct{1, []byte{'f', 'o'}})
	Test5(true, c1, c2)
	Test6(1+2i, 2+1i)
	Test7(func(i int) bool { return true })
	Test8(interface{}(2), p, up)

	fmt.Println("END")
}
