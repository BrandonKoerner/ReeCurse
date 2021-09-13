package compiler

import (
	"io"

	. "github.com/ReewassSquared/ReeCurse/compiler/lexer"
	. "github.com/ReewassSquared/ReeCurse/compiler/parser"
)

func Compile(rin io.Reader, rout io.Writer) {
	p := &ReeParser{}
	p.Parse(rin)
	
}
