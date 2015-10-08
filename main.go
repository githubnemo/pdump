package main

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// Parses the second call's parameters in a stack trace of the form:
//
// goroutine 1 [running]:
// main.printInputs(0x4c4c60, 0x539038)
//	/.../go/src/debug/main.go:16 +0xe0
// main.Test1(0x2)
//	/.../go/src/debug/main.go:23
//
func parseParams(st string) (string, []uintptr) {

	line := 1
	start, stop := 0, 0
	for i, c := range st {
		if c == '\n' {
			line++
		}
		if line == 4 && c == '\n' {
			start = i + 1
		}
		if line == 5 && c == '\n' {
			stop = i
		}
	}

	call := st[start:stop]
	fname := call[0:strings.IndexByte(call, '(')]
	param := call[strings.IndexByte(call, '(')+1 : strings.IndexByte(call, ')')]
	params := strings.Split(param, ", ")
	parsedParams := make([]uintptr, len(params))

	for i := range params {
		iv, err := strconv.ParseInt(params[i], 0, 64)

		if err != nil {
			panic(err.Error())
		}

		parsedParams[i] = uintptr(iv)
	}

	return fname, parsedParams
}

func fromAddress(t reflect.Type, addr uintptr) reflect.Value {
	return reflect.NewAt(t, unsafe.Pointer(&addr)).Elem()
}

func parameterValue(t reflect.Type, params []uintptr, pidx int) (v reflect.Value, step int) {
	switch t.Kind() {
	case reflect.Int64:
	case reflect.Int:
		// Just use the value from the stack
		v = reflect.ValueOf(params[pidx])
		step = 1
	case reflect.Float32:
		v = reflect.ValueOf(math.Float32frombits(uint32(params[pidx])))
		step = 1
	case reflect.Float64:
		v = reflect.ValueOf(math.Float64frombits(uint64(params[pidx])))
		step = 1
	case reflect.Slice:
		// create []T pointing to slice content
		data := reflect.ArrayOf(int(params[pidx+2]), t.Elem())
		svp := reflect.NewAt(data, unsafe.Pointer(params[pidx]))
		v = svp.Elem()
		step = 3
	case reflect.String:
		v = fromAddress(t, params[pidx])
		step = 2
	case reflect.Map:
		// points to hmap struct
		v = fromAddress(t, params[pidx])
		step = 1
	}
	return
}

func PrintInputs(fn interface{}) {
	v := reflect.ValueOf(fn)
	vt := v.Type()
	b := make([]byte, 500)

	if v.Kind() != reflect.Func {
		return
	}

	runtime.Stack(b, false)

	name, params := parseParams(string(b))
	pidx := 0

	fmt.Print(name + "(")
	for i := 0; i < vt.NumIn(); i++ {
		v, step := parameterValue(vt.In(i), params, pidx)
		pidx += step
		fmt.Print(v, ",")
	}
	fmt.Println(")")
}

func PrintOutputs(fn interface{}) {
	v := reflect.ValueOf(fn)
	vt := v.Type()
	b := make([]byte, 500)

	if v.Kind() != reflect.Func {
		return
	}

	runtime.Stack(b, false)

	name, params := parseParams(string(b))
	pidx := vt.NumIn()

	for i := 0; i < vt.NumOut(); i++ {
		v, step := parameterValue(vt.Out(i), params, pidx)
		pidx += step
		fmt.Print(v, ",")
	}
	fmt.Print(" = ", name, "()")
}

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

func main() {
	b := []byte{'A', 'B', 'C'}
	m := map[string]int{"foo": 3, "bar": 7}
	m["bar"] += 2
	s := "AAAA"
	Test1(2, b, 9, m)
	Test2(3.14, s)
	Test3(42)
}
