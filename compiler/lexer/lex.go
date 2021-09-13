package lexer

import (
	"fmt"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type ReeLexer struct {
	l, c    int       //line and column numbers
	b, r, e int       //used for buffer mechanics
	buf     []byte    //buffer :)
	scan    io.Reader //scanner
	ch      rune      //most recently received rune
	chw     int       //width, in bytes, of character (ReeLexer.ch)
	ioerr   error     //possible io-related error received
	bsize   uint      //size of buffer
	eof     bool      //end of file
	crash   bool      //set only if fatal error encountered during lexing
	mode    lexmode
	Tok     Token
	mstack  []reemodes
}

type reemodes struct {
	mode  lexmode
	depth int
}

type lexmode uint

const (
	LEXMODE_NORMAL lexmode = iota
	LEXMODE_QUOTE
	LEXMODE_QUASIQUOTE
)

const sentinel = utf8.RuneSelf
const LexBufferMin = 12
const LexBufferMax = 20
const ReadCountMax = 10

func (l *ReeLexer) Init(r io.Reader) {
	l.scan = r
	l.l, l.c, l.b, l.r, l.e = 0, 0, -1, 0, 0
	l.buf = make([]byte, 1<<LexBufferMin)
	l.buf[0] = sentinel
	l.ch = ' '
	l.chw = -1
	l.bsize = LexBufferMin
	l.mode = LEXMODE_NORMAL

	l.mstack = make([]reemodes, 4)
	l.mstack = append(l.mstack, reemodes{mode: LEXMODE_NORMAL, depth: 0})
}

func (l *ReeLexer) Next() *Token {
	return l.next()
}

func (l ReeLexer) Line() int {
	return l.l
}

func (l ReeLexer) Column() int {
	return l.c
}

func (_lexmode lexmode) String() string {
	switch _lexmode {
	case LEXMODE_NORMAL:
		return "normal"
	case LEXMODE_QUOTE:
		return "quote"
	case LEXMODE_QUASIQUOTE:
		return "quasiquote"
	default:
		return "<unk>"
	}
}

/* DEBUG */
func (l ReeLexer) printMstack() {
	fmt.Println("foo")
	for _, mode := range l.mstack {
		fmt.Printf("%s: %d\n", mode.mode.String(), mode.depth)
	}
}

func (l *ReeLexer) Errorf(msg string) {
	l.errorf(msg)
}

func (l *ReeLexer) start() { l.b = l.r - l.chw }
func (l *ReeLexer) stop() {
	l.b = -1
}
func (l *ReeLexer) segment() []byte {
	return l.buf[l.b : l.r-l.chw]
}

func (l *ReeLexer) errorf(msg string) {
	fmt.Printf("[%d:%d] %s\n", l.l+1, l.c+1, msg)
	//panic("")
}

func (l *ReeLexer) rewind() {
	// ok to verify precondition - rewind is rarely called
	if l.b < 0 {
		panic("no active segment")
	}
	l.c -= l.r - l.b
	l.r = l.b
	l.nextch()
}

func (l *ReeLexer) nextch() {
redo:
	l.c += int(l.chw)
	if l.ch == '\n' {
		l.l++
		l.c = 0
	}

	//first test for ASCII
	if l.ch = rune(l.buf[l.r]); l.ch < sentinel {
		l.r++
		l.chw = 1
		if l.ch == 0 {
			l.errorf("NUL")
			goto redo
		}
		return
	}

	for l.e-l.r < utf8.UTFMax && !utf8.FullRune(l.buf[l.r:l.e]) && l.ioerr == nil {
		l.fill()
	}

	//EOF
	if l.r == l.e || l.ioerr == io.EOF {
		if l.ioerr != io.EOF {
			l.errorf("IO ProgramError: " + l.ioerr.Error())
			l.ioerr = nil
		}
		l.ch = -1
		l.chw = 0
		return
	}

	l.ch, l.chw = utf8.DecodeRune(l.buf[l.r:l.e])
	l.r += l.chw
	if l.ch == utf8.RuneError && l.chw == 1 {
		l.errorf("invalid UTF-8 encoding!")
		goto redo
	}

	//WATCH OUT FOR BOM
	if l.ch == 0xfeff {
		if l.l > 0 || l.c > 0 {
			l.errorf("invalid UFT-8 byte-order mark in middle of file")
		}
		goto redo
	}
}

func (l *ReeLexer) EOF() bool {
	return l.ch < 0
}

func (l *ReeLexer) fill() {
	b := l.r
	if l.b >= 0 {
		b = l.b
		l.b = 0
	}
	content := l.buf[b:l.e]
	if len(content)*2 > len(l.buf) {
		l.bsize++
		if l.bsize > LexBufferMax {
			l.bsize = LexBufferMax
		}
		l.buf = make([]byte, 1<<l.bsize)
		copy(l.buf, content)
	} else if b > 0 {
		copy(l.buf, content)
	}
	l.r -= b
	l.e -= b

	for i := 0; i < ReadCountMax; i++ {
		var n int
		n, l.ioerr = l.scan.Read(l.buf[l.e : len(l.buf)-1])
		if n < 0 {
			panic("negative read!") //invalid io.Reader
		}
		if n > 0 || l.ioerr != nil {
			l.e += n
			l.buf[l.e] = sentinel
			return
		}
	}

	l.buf[l.e] = sentinel
	l.ioerr = io.ErrNoProgress
}

func (l *ReeLexer) next() *Token {
	switch l.mode {
	case LEXMODE_NORMAL:
		l.nextn()
		break
	case LEXMODE_QUOTE:
		l.nextq()
		break
	case LEXMODE_QUASIQUOTE:
		l.nextqq()
		break
	default:
		l.errorf("unknown mode encountered")
	}

	var tok Token
	tok = l.Tok
	return &tok
}

func (l *ReeLexer) nextn() {
redonextn:
	l.stop()
	for l.ch != -1 && (whitespace(l.ch) || l.ch == 0) {
		l.nextch()
	}
	if l.ch < 0 {
		l.Tok = l.makeToken(TOK_EOF, "")
		return
	}
	l.start()

	switch l.ch {
	case '(':
		l.nextch()
		if l.ch == ')' {
			l.nextch()
			l.Tok = l.makeToken(TOK_EMPTY, "")
			break
		}
		l.Tok = l.makeToken(TOK_LPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth++
		break
	case '[':
		l.nextch()
		l.Tok = l.makeToken(TOK_LPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth++
		break
	case ')', ']':
		l.nextch()
		l.Tok = l.makeToken(TOK_RPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth--
		if l.mstack[len(l.mstack)-1].depth <= 0 {
			/* end of current mode. revert. */
			if len(l.mstack) == 1 {
				break // root of stack; ignore and don't pop
			} else {
				l.mstack = l.mstack[:len(l.mstack)-1]   // pop
				l.mode = l.mstack[len(l.mstack)-1].mode // reset mode
				return                                  // we immediately return
			}
		}
		break
	case '+':
		l.nextch()
		if digit(l.ch) {
			l.number()
			break
		}
		l.Tok = l.makeOp(OP_ADD, "")
		break
	case '-':
		l.nextch()
		if digit(l.ch) {
			l.number()
			break
		}
		l.Tok = l.makeOp(OP_SUB, "")
		break
	case '*':
		l.nextch()
		l.Tok = l.makeOp(OP_MUL, "")
		break
	case '/':
		l.nextch()
		l.Tok = l.makeOp(OP_DIV, "")
		break
	case '"':
		l.nextch()
		l.qstring()
		break
	case '>':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.Tok = l.makeOp(OP_GTEQ, "")
			break
		}
		l.Tok = l.makeOp(OP_GT, "")
		break
	case '<':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.Tok = l.makeOp(OP_LTEQ, "")
			break
		}
		l.Tok = l.makeOp(OP_LT, "")
		break
	case '=':
		l.nextch()
		l.Tok = l.makeOp(OP_EQ, "")
		break
	case '~':
		l.nextch()
		l.Tok = l.makeOp(OP_NEQ, "")
		break
	case '\'':
		l.nextch()
		l.Tok = l.makeOp(OP_QUOTE, "")
		l.mode = LEXMODE_QUOTE
		l.mstack = append(l.mstack, reemodes{mode: l.mode, depth: 0})
		return // in this case we immediately return
	case '`':
		l.nextch()
		l.Tok = l.makeOp(OP_QUASIQUOTE, "")
		l.mode = LEXMODE_QUASIQUOTE
		l.mstack = append(l.mstack, reemodes{mode: l.mode, depth: 0})
		return // in this case we immediately return
	case '?':
		l.nextch()
		if whitespace(l.ch) {
			l.Tok = l.makeOp(OP_QUESTION, "")
			break
		}
		l.ident()
		break
	case ',':
		l.nextch()
		if l.ch == '@' {
			l.nextch()
			l.Tok = l.makeOp(OP_UNQUOTESPLICE, "")
			break
		}
		l.Tok = l.makeOp(OP_UNQUOTE, "")
		break
	case '#':
		l.nextch()
		if l.ch == 't' {
			l.nextch()
			l.Tok = l.makeToken(TOK_TRUE, "")
			break
		} else if l.ch == 'f' {
			l.nextch()
			l.Tok = l.makeToken(TOK_FALSE, "")
			break
		} else if l.ch == ';' {
			l.nextch()
			l.Tok = l.makeToken(TOK_SUPRESS, "")
			break
		} else if l.ch == '\\' {
			/* literal character */
			l.start()
			l.nextch()
			l.char()
			break
		} else if l.ch == '!' {
			l.nextch()
			l.shebang()
			return // immediately return. This should cause errors if not in top level.
		} else {
			for l.ch != -1 && !whitespace(l.ch) {
				l.nextch()
			}
			val := string(l.segment())
			l.errorf(fmt.Sprintf("unknown value after octothorp: %s", val))
			goto redonextn
		}
	default:
		if digit(l.ch) {
			l.number()
		} else {
			l.ident()
		}
	}

	if l.mstack[len(l.mstack)-1].depth <= 0 {
		/* break! */
		if len(l.mstack) == 1 {
			// root of stack; ignore and don't pop
		} else {
			fmt.Println("DEBUG: end of nextn; popping stack?")
			l.mstack = l.mstack[:len(l.mstack)-1]   // pop
			l.mode = l.mstack[len(l.mstack)-1].mode // reset mode
			return                                  // we immediately return
		}
	}
}

/**
 * Numbers are one of the following:
 * 0[xX][0-9A-Fa-f]+
 * 0[0-7]+
 * 00[01]+
 * [1-9][0-9]*
 *
 * Numbers will eventually also allow for floating point parsing.
 * For now this is disabled/unimplemented.
 *
 * Numbers can also include a sign (+/-).
 */
func (l *ReeLexer) number() {
	var val int64
	var err error
	if l.ch == '0' {
		/* octal or hex */
		l.nextch()
		if l.ch == 'x' || l.ch == 'X' {
			l.nextch()
			for l.ch != -1 && hexdigit(l.ch) {
				l.nextch()
			}
			val, err = strconv.ParseInt(string(l.segment()), 16, 64)
			if err != nil {
				l.errorf(fmt.Sprintf("invalid hexadecimal integer literal: %s", string(l.segment())))
			}
		} else if l.ch == '0' {
			/* MAKE NOTE: SIGN IS IGNORED FOR BINARY NUMBERS! */
			l.nextch()
			for l.ch != -1 && (l.ch == '0' || l.ch == '1') {
				val = 2*val + int64(l.ch-'0')
				l.nextch()
			}
		} else {
			for l.ch != -1 && digit(l.ch) {
				l.nextch()
			}
			val, err = strconv.ParseInt(string(l.segment()), 8, 64)
			if err != nil {
				l.errorf(fmt.Sprintf("invalid octal integer literal: %s", string(l.segment())))
			}
		}
	} else {
		for l.ch != -1 && digit(l.ch) {
			l.nextch()
		}
		val, err = strconv.ParseInt(string(l.segment()), 10, 64)
		if err != nil {
			l.errorf(fmt.Sprintf("invalid octal integer literal: %s", string(l.segment())))
		}
	}
	l.Tok = l.makeInt(val)
	return
}

func (l *ReeLexer) qstring() {
	l.nextch()

	for {
		if l.ch == '"' {
			l.nextch()
			break
		}
		if l.ch == '\\' {
			l.nextch()

			if !l.IsEscape('"') {
				continue
				//nothing? Empty Branch?
			}
			l.nextch()
			continue
		}
		if l.ch == '\n' {
			// to crash, or not to crash... that is the question...
			l.nextch()
		}
		if l.ch < 0 {
			l.errorf("string not terminated")
			break
		}
		l.nextch()
	}

	val, err := strconv.Unquote(string(l.segment()))
	if err != nil {
		l.errorf("invalid string literal: " + err.Error())
	}
	l.Tok = l.makeToken(TOK_LITSTR, val)
}

func (l *ReeLexer) IsEscape(quote rune) bool {
	var n int
	var base, max uint32

	switch l.ch {
	case quote, 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\':
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
		//TODO <-- find way to trigger this condition in strings test
	case 'x':
		n, base, max = 2, 16, 255
	case 'u':
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		n, base, max = 8, 16, unicode.MaxRune
	default:
		if l.ch < 0 {
			return true // complain in caller about EOF
		}
		l.errorf("unknown escape")
		return false
	}

	var x uint32
	for i := n; i > 0; i-- {
		if l.ch < 0 {
			return true // complain in caller about EOF
		}
		d := base
		if digit(l.ch) {
			d = uint32(l.ch) - '0'
		} else if 'a' <= lower(l.ch) && lower(l.ch) <= 'f' {
			d = uint32(lower(l.ch)) - 'a' + 10
		}
		if d >= base {
			l.errorf(fmt.Sprintf("invalid character %q in %s escape", l.ch, baseName(int(base))))
			return false
		}
		// d < base
		x = x*base + d
	}

	if x > max && base == 8 {
		l.errorf(fmt.Sprintf("octal escape value %d > 255", x))
		return false
	}

	if x > max || 0xD800 <= x && x < 0xE000 /* surrogate range */ {
		l.errorf(fmt.Sprintf("escape is invalid Unicode code point %#U", x))
		return false
	}

	return true
}

func (l *ReeLexer) shebang() {
	for l.ch != -1 && l.ch != '\n' {
		l.nextch()
	}
	l.Tok = l.makeToken(TOK_SHEBANG, string(l.segment()))
}

func (l *ReeLexer) ident() {
	for l.ch != -1 && !endseq(l.ch) {
		l.nextch()
	}
	if val, ok := keywords[string(l.segment())]; ok {
		l.Tok = l.makeKeyword(val, string(l.segment()))
		return
	}
	l.Tok = l.makeToken(TOK_IDENT, string(l.segment()))
}

func baseName(base int) string {
	switch base {
	case 2:
		return "binary"
	case 8:
		return "octal"
	case 10:
		return "decimal"
	case 16:
		return "hexadecimal"
	}
	panic("invalid base")
}

func (l *ReeLexer) char() {
	/* expect any unicode letter or unicode sequence. */
	for l.ch != -1 && !endseq(l.ch) {
		l.nextch()
	}

	scval := string(l.segment())
	var val rune
	var err error
	if len(scval) <= 1 {
		l.errorf("invalid character encountered")
		return
	}
	if scval[1] != 'u' && len(scval) == 2 {
		val = l.ch //should be most recent value
	} else {
		val, _, _, err = strconv.UnquoteChar(scval, '"')
	}
	if err != nil {
		l.errorf("unexpected character parsing error: " + err.Error())
	}
	l.Tok = l.makeRune(val)
}

func endseq(ch rune) bool {
	return whitespace(ch) || ch == '[' || ch == ']' || ch == '(' || ch == ')' || ch == '#' || ch == '`' || ch == '\'' || ch == '"'
}

func (l ReeLexer) makeToken(tok ReeToken, val string) Token {
	return Token{L: l.l, C: l.c, Tok: tok, Value: val}
}

func (l ReeLexer) makeInt(val int64) Token {
	return Token{L: l.l, C: l.c, Tok: TOK_LITINT, IVal: val}
}

func (l ReeLexer) makeRune(r rune) Token {
	return Token{L: l.l, C: l.c, Tok: TOK_LITCHAR, CVal: r}
}

func (l ReeLexer) makeKeyword(key ReeToken, val string) Token {
	return Token{L: l.l, C: l.c, Tok: key, Value: val}
}

func (l ReeLexer) makeOp(op ReeToken, val string) Token {
	return Token{L: l.l, C: l.c, Tok: op, Value: val}
}

func whitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func digit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func hexdigit(ch rune) bool { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

func lower(ch rune) rune { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter

func (l *ReeLexer) nextq() {
	/**
	 * ([  lparen
	 * ])  rparen
	 * .   period
	 *
	 * otherwise symbol
	 * symbol types:
	 *  number
	 *  string
	 *  true/false
	 *  char
	 *  empty
	 */

	l.stop()
	for l.ch != -1 && (whitespace(l.ch) || l.ch == 0) {
		l.nextch()
	}
	if l.ch < 0 {
		l.Tok = l.makeToken(TOK_EOF, "")
		return
	}
	l.start()

	switch l.ch {
	case '(':
		l.nextch()
		if l.ch == ')' {
			l.nextch()
			l.Tok = l.makeToken(SYM_EMPTY, "")
			break
		}
		l.Tok = l.makeToken(TOK_LPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth++
		break
	case '[':
		l.nextch()
		l.Tok = l.makeToken(TOK_LPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth++
		break
	case ')', ']':
		l.nextch()
		l.Tok = l.makeToken(TOK_RPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth--
		if l.mstack[len(l.mstack)-1].depth <= 0 {
			/* end of current mode. revert. */
			if len(l.mstack) == 1 {
				break // root of stack; ignore and don't pop
			} else {
				l.mstack = l.mstack[:len(l.mstack)-1]   // pop
				l.mode = l.mstack[len(l.mstack)-1].mode // reset mode
				return                                  // we immediately return
			}
		}
		break
	case '.':
		l.nextch()
		l.Tok = l.makeToken(TOK_PERIOD, "")
		break
	case '"':
		l.nextch()
		l.qstring()
		/* qstring builds string literal. change to sym. */
		l.Tok.Tok = SYM_LITSTR
		break
	case '#':
		l.nextch()
		if l.ch == 't' {
			l.nextch()
			l.Tok = l.makeToken(SYM_TRUE, "")
			break
		} else if l.ch == 'f' {
			l.nextch()
			l.Tok = l.makeToken(SYM_FALSE, "")
			break
		} else if l.ch == '\\' {
			/* literal character */
			l.start()
			l.nextch()
			l.char()
			l.Tok.Tok = SYM_LITCHAR
			break
		}
	default:
		if digit(l.ch) {
			l.number()
			l.Tok.Tok = SYM_LITINT
		} else {
			l.ident()
			l.Tok.Tok = TOK_SYMBOL
		}
	}

	if l.mstack[len(l.mstack)-1].depth <= 0 {
		/* break! */
		if len(l.mstack) == 1 {
			// root of stack; ignore and don't pop
		} else {
			l.mstack = l.mstack[:len(l.mstack)-1]   // pop
			l.mode = l.mstack[len(l.mstack)-1].mode // reset mode
			return                                  // we immediately return
		}
	}
}

func (l *ReeLexer) nextqq() {
	//l.printMstack()
	l.stop()
	for l.ch != -1 && (whitespace(l.ch) || l.ch == 0) {
		l.nextch()
	}
	if l.ch < 0 {
		l.Tok = l.makeToken(TOK_EOF, "")
		return
	}
	l.start()

	switch l.ch {
	case '(':
		l.nextch()
		if l.ch == ')' {
			l.nextch()
			l.Tok = l.makeToken(SYM_EMPTY, "")
			break
		}
		l.Tok = l.makeToken(TOK_LPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth++
		break
	case '[':
		l.nextch()
		l.Tok = l.makeToken(TOK_LPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth++
		break
	case ')', ']':
		l.nextch()
		l.Tok = l.makeToken(TOK_RPAREN, "")
		/* increment */
		l.mstack[len(l.mstack)-1].depth--
		if l.mstack[len(l.mstack)-1].depth <= 0 {
			/* end of current mode. revert. */
			if len(l.mstack) == 1 {
				break // root of stack; ignore and don't pop
			} else {
				l.mstack = l.mstack[:len(l.mstack)-1]   // pop
				l.mode = l.mstack[len(l.mstack)-1].mode // reset mode
				return                                  // we immediately return
			}
		}
		break
	case '.':
		l.nextch()
		l.Tok = l.makeToken(TOK_PERIOD, "")
		break
	case '"':
		l.nextch()
		l.qstring()
		/* qstring builds string literal. change to sym. */
		l.Tok.Tok = SYM_LITSTR
		break
	case ',':
		l.nextch()
		if l.ch == '@' {
			l.nextch()
			l.Tok = l.makeOp(OP_UNQUOTESPLICE, "")
			l.mode = LEXMODE_NORMAL
			l.mstack = append(l.mstack, reemodes{mode: l.mode, depth: 0})
			return
		}
		l.Tok = l.makeOp(OP_UNQUOTE, "")
		l.mode = LEXMODE_NORMAL
		l.mstack = append(l.mstack, reemodes{mode: l.mode, depth: 0})
		return
	case '#':
		l.nextch()
		if l.ch == 't' {
			l.nextch()
			l.Tok = l.makeToken(SYM_TRUE, "")
			break
		} else if l.ch == 'f' {
			l.nextch()
			l.Tok = l.makeToken(SYM_FALSE, "")
			break
		} else if l.ch == '\\' {
			/* literal character */
			l.start()
			l.nextch()
			l.char()
			l.Tok.Tok = SYM_LITCHAR
			break
		}
	default:
		if digit(l.ch) {
			l.number()
			l.Tok.Tok = SYM_LITINT
		} else {
			l.ident()
			l.Tok.Tok = TOK_SYMBOL
		}
	}

	if l.mstack[len(l.mstack)-1].depth <= 0 {
		/* break! */
		if len(l.mstack) == 1 {
			// root of stack; ignore and don't pop
		} else {
			l.mstack = l.mstack[:len(l.mstack)-1]   // pop
			l.mode = l.mstack[len(l.mstack)-1].mode // reset mode
			return                                  // we immediately return
		}
	}
}
