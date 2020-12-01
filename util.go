package sdiffer

import (
	"fmt"
	"reflect"
	"strings"
)

func isStringBlank(str string) bool {
	str = strings.TrimSpace(str)
	return len(str) == 0
}

func mustSuccess(fn func() error) {
	if err := fn(); err != nil {
		panic(err)
	}
}

func allowPanic(fn func()) (isPanicked bool) {
	defer func() {
		if r := recover(); r != nil {
			isPanicked = true
			fmt.Println(r)
		}
	}()
	fn()
	return
}

func iF(condition bool, a, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
}

func concat(strList ...string) string {
	builder := &strings.Builder{}
	for _, str := range strList {
		builder.WriteString(str)
	}
	return builder.String()
}

func toString(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copySliceValue(sv reflect.Value) reflect.Value {
	length := sv.Len()
	copiedSv := reflect.MakeSlice(sv.Type(), length, length)
	for i := 0; i < length; i++ {
		copiedSv.Index(i).Set(sv.Index(i))
	}
	return copiedSv
}