package marshal

import (
)

const (
	MARSHAL_VERSION     = 3
	TYPE_NULL           = '0'
	TYPE_NONE           = 'N'
	TYPE_FALSE          = 'F'
	TYPE_TRUE           = 'T'
	TYPE_STOPITER       = 'S'
	TYPE_ELLIPSIS       = '.'
	TYPE_INT            = 'i'
	TYPE_FLOAT          = 'f'
	TYPE_BINARY_FLOAT   = 'g'
	TYPE_COMPLEX        = 'x'
	TYPE_BINARY_COMPLEX = 'y'
	TYPE_LONG           = 'l'
	TYPE_STRING         = 's'
	TYPE_INTERNED       = 't'
	TYPE_REF            = 'r'
	TYPE_TUPLE          = '('
	TYPE_LIST           = '['
	TYPE_DICT           = '{'
	TYPE_CODE           = 'c'
	TYPE_UNICODE        = 'u'
	TYPE_UNKNOWN        = '?'
	TYPE_SET            = '<'
	TYPE_FROZENSET      = '>'
	FLAG_REF            = 0x80 // with a type, add obj to index
	SIZE32_MAX          = 0x7FFFFFFF

	TYPE_ASCII                = 'a'
	TYPE_ASCII_INTERNED       = 'A'
	TYPE_SMALL_TUPLE          = ')'
	TYPE_SHORT_ASCII          = 'z'
	TYPE_SHORT_ASCII_INTERNED = 'Z'

	// We assume that Python ints are stored internally in base some power of
	// 2**15; for the sake of portability we'll always read and write them in base
	// exactly 2**15.

	PyLong_MARSHAL_SHIFT = 15
	PyLong_MARSHAL_BASE  = (1 << PyLong_MARSHAL_SHIFT)
	PyLong_MARSHAL_MASK  = (PyLong_MARSHAL_BASE - 1)
)


