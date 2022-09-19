package marshal

import (
	"encoding/binary"
	// "fmt"
	"io"
)

type PyListObject struct {
	reader   io.Reader
	items    []_object
	typecode byte
}

func (plo *PyListObject) r_object() _object {
	var nItems int
	if plo.typecode == TYPE_SMALL_TUPLE {
		var size int8
		if err := binary.Read(plo.reader, binary.LittleEndian, &size); err != nil {
			panic("Failed to read SMALl_TUPLE size")
		}
		nItems = int(size)
		// fmt.Println("small_tuple, size=", nItems)
	} else {
		var size int32
		if err := binary.Read(plo.reader, binary.LittleEndian, &size); err != nil {
			panic("Failed to read size")
		}
		nItems = int(size)
		// fmt.Println("list or tuple, size=", nItems)
	}

	for i := 0; i < nItems; i++ {
		po := &PyObject{plo.reader}
		plo.items = append(plo.items, po.r_object())
	}
	return plo
}

func (plo *PyListObject) GetItems() []_object {
	return plo.items
}
