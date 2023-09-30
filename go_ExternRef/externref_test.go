package main

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

var (
	//go:embed funcs.wasm
	funcsWasm []byte
)

func TestExt(t *testing.T) {
	// Set not to print debug info
	wasmedge.SetLogErrorLevel()

	// Create VM
	var vm = wasmedge.NewVM()

	// Create module instance
	var obj = wasmedge.NewModule("extern_module")

	// Add host functions into the module instance
	var type1 = wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_ExternRef,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		}, []wasmedge.ValType{
			wasmedge.ValType_I32,
		})
	var type2 = wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_ExternRef,
			wasmedge.ValType_I32,
		}, []wasmedge.ValType{
			wasmedge.ValType_I32,
		})
	var func_add = wasmedge.NewFunction(type1, host_add, nil, 0)
	var func_mul = wasmedge.NewFunction(type1, host_mul, nil, 0)
	var func_square = wasmedge.NewFunction(type2, host_square, nil, 0)
	obj.AddFunction("add", func_add)
	obj.AddFunction("mul", func_mul)
	obj.AddFunction("square", func_square)

	// Register module instance
	vm.RegisterModule(obj)

	// Instantiate wasm
	vm.LoadWasmBuffer(funcsWasm)
	vm.Validate()
	vm.Instantiate()

	// Run
	var ref_add = wasmedge.NewExternRef(real_add)
	var ref_mul = wasmedge.NewExternRef(real_mul)
	var ref_square = wasmedge.NewExternRef(real_square)
	var res []interface{}
	var err error
	res, err = vm.Execute("call_add", ref_add, int32(1234), int32(5678))
	if err == nil {
		fmt.Println("Run call_add: 1234 + 5678 =", res[0].(int32))
	} else {
		fmt.Println("Run call_add FAILED")
	}
	res, err = vm.Execute("call_mul", ref_mul, int32(4827), int32(-31519))
	if err == nil {
		fmt.Println("Run call_mul: 4827 * (-31519) =", res[0].(int32))
	} else {
		fmt.Println("Run call_mul FAILED")
	}
	res, err = vm.Execute("call_square", ref_square, int32(1024))
	if err == nil {
		fmt.Println("Run call_square: 1024^2 =", res[0].(int32))
	} else {
		fmt.Println("Run call_square FAILED")
	}
	res, err = vm.Execute("call_add_square", ref_add, ref_square, int32(761), int32(195))
	if err == nil {
		fmt.Println("Run call_square: (761 + 195)^2 =", res[0].(int32))
	} else {
		fmt.Println("Run call_square FAILED")
	}

	ref_add.Release()
	ref_mul.Release()
	ref_square.Release()
	vm.Release()
	obj.Release()
}
