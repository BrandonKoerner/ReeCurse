package parser

import (
	"fmt"
	"io"

	. "github.com/ReewassSquared/ReeCurse/compiler/lexer"
)

type ReeParser struct {
	*ReeLexer
	Node *Node
}

func (p ReeParser) got(tok ReeToken) bool {
	return p.Tok.Tok == tok
}

func (p ReeParser) want(tok ReeToken) {
	if !p.got(tok) {
		p.Errorf(fmt.Sprintf("unexpected %s; wanted %s", p.Tok.Tok.String(), tok.String()))
	}
	p.Next()
}

func (p *ReeParser) Parse(r io.Reader) {
	p.Init(r)

	for !p.got(TOK_EOF) {
		p.ParseExpr()
	}
}

/**
 * EXPR = int | boolean | string | symbol
 *      | empty | character | variable
 *      | quote_expr | quasiquote_expr
 *		| if_expr    | let_expr   		| letrec_expr
 *		| cond_expr  | match_expr
 *		| define	 | unary_expr		| binary_expr
 *		| *ary_expr
 */
func (p *ReeParser) ParseExpr() {
	tok := p.Next()
	switch tok.Tok {
	case TOK_LITINT:
		node := p.MakeNode(NODE_INTEGER)
		node.Etype = typemap["int"]
		break
	default:
		p.Errorf("unimplemented")
	}
}

func (p *ReeParser) MakeNode(ntype Nodetype) *Node {
	return &Node{L: p.Line(), C: p.Column(), Ntype: ntype}
}
