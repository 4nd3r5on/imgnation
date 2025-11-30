package utils

import (
	"strconv"
	"unsafe"
)

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Int interface {
	Signed | Unsigned
}

type Float interface {
	~float32 | ~float64
}

type Num interface {
	Int | Float
}

func GetNumSize[T Num]() int {
	var zero T
	return int(unsafe.Sizeof(zero)) * 8
}

func ParseInt[T Int](s string) (T, error) {
	numSize := GetNumSize[T]()
	val, err := strconv.ParseInt(s, 10, numSize)
	return T(val), err
}

func ParseFloat[T Float](s string) (T, error) {
	numSize := GetNumSize[T]()
	val, err := strconv.ParseFloat(s, numSize)
	return T(val), err
}
