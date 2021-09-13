package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/ReewassSquared/ReeCurse/compiler/lexer"
)

func main() {
	src, err := ioutil.ReadFile("script.curse")
	if err != nil {
		panic("File Error")
	}

	source := string(src)
	b := bytes.NewBufferString(source)
	l := lexer.ReeLexer{}
	counter := 0
	l.Init(b)
	for !l.EOF() && counter < 100 {
		l.Next()
		if l.Tok.Tok == lexer.TOK_KEYWORD {
			//fmt.Printf("[%4d:%4d] %18s %s\n", l.Tok.L, l.Tok.C, l.Tok.Key.String(), l.Tok.Value)
		} else if l.Tok.Tok == lexer.TOK_KEYOP {
			//fmt.Printf("[%4d:%4d] %18s %s\n", l.Tok.L, l.Tok.C, l.Tok.Op.String(), l.Tok.Value)
		} else if l.Tok.Tok == lexer.TOK_LITINT {
			fmt.Printf("[%4d:%4d] %18s %d\n", l.Tok.L, l.Tok.C, l.Tok.Tok.String(), l.Tok.IVal)
		} else if l.Tok.Tok == lexer.TOK_LITCHAR {
			fmt.Printf("[%4d:%4d] %18s %c\n", l.Tok.L, l.Tok.C, l.Tok.Tok.String(), l.Tok.CVal)
		} else {
			fmt.Printf("[%4d:%4d] %18s %s\n", l.Tok.L, l.Tok.C, l.Tok.Tok.String(), l.Tok.Value)
		}
		counter++
	}
}
