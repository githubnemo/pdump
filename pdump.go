package pdump

// Dump and access input/output parameter values of surrounding functions.
//
// Example usage:
//
// Code:
//
// 	import "github.com/githubnemo/pdump"
//
// 	func Test3(in int) (int, int) {
// 		pdump.PrintInputs(Test3)
// 		defer pdump.PrintOutputs(Test3)
// 		return 3, 4
// 	}
//
// 	func main() {
// 		Test3(42)
// 	}
//
// Output:
//
// 	main.Test3(42)
// 	3,4, = main.Test3()
//

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
// main.Test1(0x2)                       <---- parsed
//	/.../go/src/debug/main.go:23
//
//	Returns the function name and the parameter values, e.g.:
//
//		("main.Test1", [0x2])
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
	case reflect.Bool:
		v = reflect.ValueOf(params[pidx]&0xFF == 1)
		step = 1
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Int:
		v = reflect.New(t).Elem()
		v.SetInt(int64(params[pidx]))
		step = 1
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
		v = reflect.New(t).Elem()
		v.SetUint(uint64(params[pidx]))
		step = 1
	case reflect.Float32:
		v = reflect.ValueOf(math.Float32frombits(uint32(params[pidx])))
		step = 1
	case reflect.Float64:
		v = reflect.ValueOf(math.Float64frombits(uint64(params[pidx])))
		step = 1
	case reflect.Complex64:
		v = fromAddress(t, params[pidx])
		step = 1
	case reflect.Complex128:
		v = fromAddress(t, params[pidx])
		step = 2
	case reflect.Array:
		v = fromAddress(t, params[pidx])
		step = 1
	case reflect.Slice:
		z := params[pidx : pidx+2]
		v = reflect.NewAt(t, unsafe.Pointer(&z[0])).Elem()
		step = 3
	case reflect.Func:
		v = fromAddress(t, params[pidx])
		step = 1
	case reflect.Interface:
		z := params[pidx : pidx+1]
		v = reflect.NewAt(t, unsafe.Pointer(&z[0])).Elem()
		step = 2
	case reflect.Ptr:
		v = fromAddress(t, params[pidx])
		step = 1
	case reflect.String:
		v = fromAddress(t, params[pidx])
		step = 2
	case reflect.Map:
		// points to hmap struct
		v = fromAddress(t, params[pidx])
		step = 1
	case reflect.Struct:
		// Determine overall step count over all fields
		os := 0
		for i := 0; i < t.NumField(); i++ {
			_, s := parameterValue(t.Field(i).Type, params, pidx+os)
			os += s
		}
		z := params[pidx : pidx+os]
		v = reflect.NewAt(t, unsafe.Pointer(&z[0])).Elem()
		step = os
	case reflect.Chan:
		v = fromAddress(t, params[pidx])
		step = 1
	case reflect.UnsafePointer:
		v = fromAddress(t, params[pidx])
		step = 1
	}
	return
}

func inputParameterValues(fn interface{}, stack []byte) (string, []reflect.Value) {
	v := reflect.ValueOf(fn)
	vt := v.Type()

	if v.Kind() != reflect.Func {
		return "", nil
	}

	name, params := parseParams(string(stack))
	pidx := 0
	pparams := make([]reflect.Value, vt.NumIn())

	for i := 0; i < vt.NumIn(); i++ {
		v, step := parameterValue(vt.In(i), params, pidx)
		pidx += step
		pparams[i] = v
	}

	return name, pparams
}

func outputParameterValues(fn interface{}, stack []byte) (string, []reflect.Value) {
	v := reflect.ValueOf(fn)
	vt := v.Type()

	if v.Kind() != reflect.Func {
		return "", nil
	}

	name, params := parseParams(string(stack))
	pidx := vt.NumIn()
	pparams := make([]reflect.Value, vt.NumOut())

	for i := 0; i < vt.NumOut(); i++ {
		v, step := parameterValue(vt.Out(i), params, pidx)
		pidx += step
		pparams[i] = v
	}
	return name, pparams
}

// Returns the input parameters of the function surrounding the call to this
// function as reflection values. It expects the function as a parameter to
// have type information about the input parameters.
func Inputs(fn interface{}) []reflect.Value {
	v := reflect.ValueOf(fn)

	if v.Kind() != reflect.Func {
		return nil
	}

	b := make([]byte, 500)
	runtime.Stack(b, false)

	_, params := inputParameterValues(fn, b)
	return params
}

// Analog of Inputs to output (return) values.
func Outputs(fn interface{}) []reflect.Value {
	v := reflect.ValueOf(fn)

	if v.Kind() != reflect.Func {
		return nil
	}

	b := make([]byte, 500)
	runtime.Stack(b, false)

	_, params := outputParameterValues(fn, b)
	return params
}

// Prints the input parameter values of the surrounding function in a human
// readable format.
func PrintInputs(fn interface{}) {
	v := reflect.ValueOf(fn)

	if v.Kind() != reflect.Func {
		return
	}

	b := make([]byte, 500)
	runtime.Stack(b, false)

	name, params := inputParameterValues(fn, b)

	fmt.Print(name, "(")

	for i, v := range params {
		if i != 0 {
			fmt.Print(",")
		}
		fmt.Printf("%#v", v)
	}

	fmt.Println(")")
}

// Prints the output parameter values of the surrounding function in a human
// readable format. Note that this will only work if the call to this function
// is deferred.
func PrintOutputs(fn interface{}) {
	v := reflect.ValueOf(fn)

	if v.Kind() != reflect.Func {
		return
	}

	b := make([]byte, 500)
	runtime.Stack(b, false)

	name, params := outputParameterValues(fn, b)

	for _, v := range params {
		fmt.Printf("%#v,", v)
	}

	fmt.Printf(" = %s()\n", name)
}
