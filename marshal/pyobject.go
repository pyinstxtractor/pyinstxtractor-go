package marshal

import (
	"encoding/binary"
	// "fmt"
	"io"
)

type PyObject struct {
	reader io.Reader
}

func (po *PyObject) r_object() _object {
	var code byte
	if err := binary.Read(po.reader, binary.LittleEndian, &code); err != nil {
		panic("Failed to read code byte")
	}

	addRef := (code & FLAG_REF) != 0
	typecode := code &^ FLAG_REF
	// fmt.Printf("%c ", typecode)
	var obj _object

	switch typecode {
	case TYPE_LIST, TYPE_TUPLE, TYPE_SMALL_TUPLE:
		obj = &PyListObject{reader: po.reader, typecode: typecode}
		obj.r_object()

	case TYPE_SHORT_ASCII, TYPE_SHORT_ASCII_INTERNED,
		TYPE_STRING, TYPE_INTERNED, TYPE_UNICODE,
		TYPE_ASCII, TYPE_ASCII_INTERNED:
		obj = &PyStringObject{reader: po.reader, typecode: typecode}
		obj.r_object()

	case TYPE_INT:
		obj = &PyIntegerObject{reader: po.reader}
		obj.r_object()

	case TYPE_REF:
		// Reference to a previous read
		var n int32
		if err := binary.Read(po.reader, binary.LittleEndian, &n); err != nil {
			panic("Failed to read TYPE_REF")
		}

		if n < 1 || int(n) >= len(_unmarshaler.refs) {
			panic("TYPE_REF out of bounds")
		}
		n -= 1
		// fmt.Println("Get ref", n)
		obj = _unmarshaler.refs[n]

	default:
		panic("Unsupported typecode: " + string(typecode))

	}
	if addRef {
		// fmt.Println("Added ref", len(_unmarshaler.refs))
		_unmarshaler.refs = append(_unmarshaler.refs, obj)
	}
	return obj
}
