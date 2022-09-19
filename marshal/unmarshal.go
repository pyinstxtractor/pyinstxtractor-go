package marshal

import (
	"fmt"
	"io"
)

type SimpleUnmarshaler struct {
	reader io.Reader
	refs []_object
}

var _unmarshaler *SimpleUnmarshaler

func NewUnmarshaler(r io.Reader) *SimpleUnmarshaler {
	_unmarshaler = &SimpleUnmarshaler{reader: r}
	return _unmarshaler
}

func (su *SimpleUnmarshaler) Unmarshal() _object {
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("Panicked during unmarshal!")
			fmt.Println(r)
        }
    }()

	pobj := PyObject{reader: su.reader}
	return pobj.r_object()
}