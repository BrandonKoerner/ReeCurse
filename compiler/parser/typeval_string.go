// Code generated by "stringer -type=TypeVal"; DO NOT EDIT.

package parser

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TYPE_UNK-0]
	_ = x[TYPE_INT-1]
	_ = x[TYPE_STRING-2]
	_ = x[TYPE_CHAR-3]
	_ = x[TYPE_BOOLEAN-4]
	_ = x[TYPE_SYMBOL-5]
	_ = x[TYPE_BOX-6]
	_ = x[TYPE_CONS-7]
	_ = x[TYPE_LIST-8]
}

const _TypeVal_name = "TYPE_UNKTYPE_INTTYPE_STRINGTYPE_CHARTYPE_BOOLEANTYPE_SYMBOLTYPE_BOXTYPE_CONSTYPE_LIST"

var _TypeVal_index = [...]uint8{0, 8, 16, 27, 36, 48, 59, 67, 76, 85}

func (i TypeVal) String() string {
	if i >= TypeVal(len(_TypeVal_index)-1) {
		return "TypeVal(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TypeVal_name[_TypeVal_index[i]:_TypeVal_index[i+1]]
}
