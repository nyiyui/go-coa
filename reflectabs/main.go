package reflect

import (
	"reflect"
	"unsafe"
)

// Value is an abstraction of reflect.Value, so programs can use other implementations.
type Value interface {
	Addr() Value
	Bool() bool
	Bytes() []byte
	Call(in []Value) []Value
	CallSlice(in []Value) []Value
	CanAddr() bool
	CanConvert(t Type) bool
	CanInterface() bool
	CanSet() bool
	Cap() int
	Close()
	Complex() complex128
	Convert(t Type) Value
	Elem() Value
	Field(i int) Value
	FieldByIndex(index []int) Value
	FieldByName(name string) Value
	FieldByNameFunc(match func(string) bool) Value
	Float() float64
	Index(i int) Value
	Int() int64
	Interface() (i interface{})
	InterfaceData() [2]uintptr
	IsNil() bool
	IsValid() bool
	IsZero() bool
	Kind() Kind
	Len() int
	MapIndex(key Value) Value
	MapKeys() []Value
	MapRange() MapIter
	Method(i int) Value
	MethodByName(name string) Value
	NumField() int
	NumMethod() int
	OverflowComplex(x complex128) bool
	OverflowFloat(x float64) bool
	OverflowInt(x int64) bool
	OverflowUint(x uint64) bool
	Pointer() uintptr
	Recv() (x Value, ok bool)
	Send(x Value)
	Set(x Value)
	SetBool(x bool)
	SetBytes(x []byte)
	SetCap(n int)
	SetComplex(x complex128)
	SetFloat(x float64)
	SetInt(x int64)
	SetLen(n int)
	SetMapIndex(key, elem Value)
	SetPointer(x unsafe.Pointer)
	SetString(x string)
	SetUint(x uint64)
	Slice(i, j int) Value
	Slice3(i, j, k int) Value
	String() string
	TryRecv() (x Value, ok bool)
	TrySend(x Value) bool
	Type() Type
	Uint() uint64
	UnsafeAddr() uintptr
}

// Type is a type alias of the reflect.Type interface
// (since reflect.Type is an interface, it does not need another abstraction.
// The purpose is to have all the types in one package.
type Type = reflect.Type

// MapIter is an abstraction of reflect.MapIter, so programs can use other implementations.
type MapIter interface {
	Key() Value
	Next() bool
	Value() Value
}

// Kind is a type alias of reflect.Kind. The purpose is to have all the types in one package.
type Kind = reflect.Kind
