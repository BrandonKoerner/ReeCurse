//go:generate go run golang.org/x/tools/cmd/stringer -type=ReeToken

package lexer

type Token struct {
	L, C  int    //line and column
	Value string //value of character
	IVal  int64  //used for integers
	CVal  rune   //used for characters
	Tok   ReeToken
}

type ReeToken uint

const (
	TOK_UNDEF ReeToken = iota
	TOK_SHEBANG
	TOK_LPAREN
	TOK_RPAREN
	TOK_LITINT
	TOK_LITNUM
	TOK_LITSTR
	TOK_IDENT
	TOK_KEYWORD
	TOK_KEYOP
	TOK_EMPTY
	TOK_TRUE
	TOK_FALSE
	TOK_LITCHAR
	TOK_SUPRESS
	SYM_LITCHAR
	SYM_LITINT
	SYM_LITSTR
	SYM_TRUE
	SYM_FALSE
	SYM_EMPTY
	TOK_SYMBOL
	TOK_PERIOD

	KEY_UNDEF
	KEY_TYPE
	KEY_LET
	KEY_LETREC
	KEY_IF
	KEY_DEFINE
	KEY_COND
	KEY_MATCH
	KEY_ELSE
	KEY_LAMBDA

	OP_UNDEF
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_ZERO
	OP_ABS
	OP_GT
	OP_GTEQ
	OP_LTEQ
	OP_LT
	OP_INC
	OP_DEC
	OP_EQ
	OP_NEQ
	OP_PRINT
	OP_BOX
	OP_UNBOX
	OP_CONS
	OP_CAR
	OP_CDR
	OP_QUOTE
	OP_QUASIQUOTE
	OP_UNQUOTE
	OP_UNQUOTESPLICE
	OP_CHECKTYPE
	OP_MOD
	OP_NOT
	OP_BITAND
	OP_BITOR
	OP_BITXOR
	OP_BITSHL
	OP_BITSHR
	OP_QUESTION

	TOK_EOF
)

var keywords map[string]ReeToken = map[string]ReeToken{
	"type":   KEY_TYPE,
	"let":    KEY_LET,
	"let*":   KEY_LETREC,
	"if":     KEY_IF,
	"cond":   KEY_COND,
	"else":   KEY_ELSE,
	"define": KEY_DEFINE,
	"lambda": KEY_LAMBDA,
	"λ":      KEY_LAMBDA,
	"match":  KEY_MATCH,
}
