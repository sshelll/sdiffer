package sdiffer

import (
	"bytes"
	"fmt"
)

type bufferF struct {
	bytes.Buffer
}

func newBufferF() *bufferF {
	return &bufferF{
		bytes.Buffer{},
	}
}

func (bff *bufferF) sprintf(format string, args ...interface{}) {
	mustSuccess(func() (err error) {
		_, err = bff.WriteString(fmt.Sprintf(format, args...))
		return
	})
}