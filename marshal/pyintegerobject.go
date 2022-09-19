package marshal

import (
	"encoding/binary"
	// "fmt"
	"io"
)

type PyIntegerObject struct {
	reader io.Reader
	value int32
}

func (pio *PyIntegerObject) r_object() _object {
	if err := binary.Read(pio.reader, binary.LittleEndian, &pio.value); err != nil {
		panic("Failed to read integer object")
	}

	// fmt.Println("integer_object, value=", pio.value)
	return pio
}

func (pio *PyIntegerObject) GetValue() int {
	return int(pio.value)
}