//go:generate go run golang.org/x/tools/cmd/stringer -type=TypeVal
//go:generate go run golang.org/x/tools/cmd/stringer -type=Nodetype

package parser

import (
	. "github.com/ReewassSquared/ReeCurse/compiler/lexer"
)

type Node struct {
	L, C  int
	Ntype Nodetype

	Left, Right *Node

	Op      ReeToken
	Etype   *ReeType
	Supress bool

	Nodes []*Node

	/* defines and other non-return expressions can be in here. */
	/* Node is treated as closure if definitions are mutable */
	Scope []*Node
}

type Nodetype uint
type TypeVal uint

/**
 * Types.
 * Typing can be native, tree-like or even parametric, AND recursive.
 * Parametric types have
 *
 *
 *
 *
 *
 *
 *
 *
 *
 */
type ReeType struct {
	Val    TypeVal
	Native bool
	Types  []*ReeType //saves on memory if not needed :)
	Name   string     //sometimes not needed.
	Params []*ReeType //
}

var typemap map[string]*ReeType = map[string]*ReeType{}

const (
	TYPE_UNK TypeVal = iota
	TYPE_INT
	TYPE_STRING
	TYPE_CHAR
	TYPE_BOOLEAN
	TYPE_SYMBOL
	TYPE_BOX
	TYPE_CONS
	TYPE_LIST
	TYPE_CUSTOM
)

const (
	NODE_UNDEF Nodetype = iota
	NODE_INTEGER
	NODE_STRING
	NODE_BOOLEAN
	NODE_UNARY
	NODE_BINARY
	NODE_IF
	NODE_COND
	NODE_LET
	NODE_CLAUSE
	NODE_BIND
	NODE_VARIABLE
	NODE_EMPTY
	NODE_DEFINE
	NODE_QUOTE
	NODE_MATCH
	NODE_MATCHCLAUSE
)

func addNativeType(typ TypeVal, name string) {
	typemap[name] = &ReeType{Val: typ, Native: true, Name: name}
}

func init() {
	/* populate the typemap with builtin types */
	addNativeType(TYPE_INT, "int")
	addNativeType(TYPE_BOOLEAN, "bool")
}
