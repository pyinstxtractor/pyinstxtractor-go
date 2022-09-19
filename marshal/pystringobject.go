package marshal

import (
	"encoding/binary"
	// "fmt"
	"io"
)

type PyStringObject struct {
	reader   io.Reader
	value    string
	typecode byte
}

func (pso *PyStringObject) r_object() _object {
	var length int

	switch pso.typecode {
	case TYPE_SHORT_ASCII, TYPE_SHORT_ASCII_INTERNED:
		var size uint8
		err := binary.Read(pso.reader, binary.LittleEndian, &size)
		if err != nil {
			return nil
		}
		length = int(size)

	case TYPE_STRING, TYPE_INTERNED, TYPE_UNICODE,
		TYPE_ASCII, TYPE_ASCII_INTERNED:
		var size int32
		err := binary.Read(pso.reader, binary.LittleEndian, &size)
		if err != nil {
			return nil
		}
		length = int(size)
	}

	buf := make([]byte, length)
	_, err := io.ReadFull(pso.reader, buf)
	if err != nil {
		return nil
	}

	pso.value = string(buf)
	// fmt.Println("string_object, value=", pso.value)
	return pso
}

func(pso *PyStringObject) GetString() string {
	return pso.value
}
