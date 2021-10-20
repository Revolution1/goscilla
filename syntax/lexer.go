package syntax

import (
	"bytes"
	"fmt"
	"github.com/rhysd/locerr"
	"goscilla/token"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type stateFn func(*Lexer) stateFn

const eof = -1

// Lexer instance which contains lexing states.
type Lexer struct {
	state            stateFn
	start            locerr.Pos
	current          locerr.Pos
	currentPosString string
	src              *locerr.Source
	input            *bytes.Reader
	Tokens           chan token.Token
	top              rune
	eof              bool
	// Function called when error occurs.
	// By default it outputs an error to stderr.
	Error func(msg string, pos locerr.Pos)
}

// NewLexer creates new Lexer instance.
func NewLexer(src *locerr.Source) *Lexer {
	start := locerr.Pos{
		Offset: 0,
		Line:   1,
		Column: 1,
		File:   src,
	}
	return &Lexer{
		state:   lex,
		start:   start,
		current: start,
		input:   bytes.NewReader(src.Code),
		src:     src,
		Tokens:  make(chan token.Token),
		Error:   nil,
	}
}

// Lex starts lexing. Lexed tokens will be queued into channel in lexer.
func (l *Lexer) Lex() {
	// Set top to peek current rune
	l.forward()
	for l.state != nil {
		l.state = l.state(l)
	}
}

func (l *Lexer) emitIdent(ident string) {
	for idx, kw := range token.KeywordTable {
		if ident == kw {
			l.emit(token.Kind(idx))
			return
		}
	}

	for idx, adt := range token.BuiltinADTTable {
		if ident == adt {
			l.emit(token.Kind(idx))
			return
		}
	}

	switch {
	case ident[0] == '_':
		l.emit(token.SPID)
		return
	case 'A' <= ident[0] && ident[0] <= 'Z':
		l.emit(token.CID)
	default:
		l.emit(token.ID)
	}
}

func (l *Lexer) emitPrimeType(ident string) bool {
	switch ident {
	case "Int32", "Int64", "Int128", "Int256":
	case "Uint32", "Uint64", "Uint128", "Uint256":
		l.emit(token.INT_TYPE)
		return true
	case "Event":
		l.emit(token.EVENT_TYPE)
		return true
	case "Message":
		l.emit(token.MESSAGE_TYPE)
		return true
	case "BNum":
		l.emit(token.BNUM_TYPE)
		return true
	case "String":
		l.emit(token.STRING_TYPE)
		return true
	}
	if strings.HasPrefix(ident, "ByStr") {
		l.emit(token.BYSTR_TYPE)
		return true
	}
	return false
}

func (l *Lexer) emit(kind token.Kind) {
	tok := token.Token{
		Kind:  kind,
		Start: l.start,
		End:   l.current,
		File:  l.src,
	}
	l.Tokens <- tok
	l.start = l.current
	if traceLexerTokenPos {
		line := l.getCurrentPosString(tok.Start.Offset - tok.End.Offset)
		_, _ = fmt.Fprintf(os.Stderr, "%s (%d:%d)\n", line, tok.Start.Line, tok.Start.Column)
	}
	if traceLexerToken {
		_, _ = fmt.Fprintf(os.Stderr, "Token: %s\n", tok.String())
	}
}
func (l *Lexer) emitIllegal(reason string) {
	l.errmsg(reason)
	tok := token.Token{
		Kind:  token.ILLEGAL,
		Start: l.start,
		End:   l.current,
		File:  l.src,
	}
	l.Tokens <- tok
	l.start = l.current
	if traceLexerTokenPos {
		line := l.getCurrentPosString(tok.Start.Offset - tok.End.Offset)
		_, _ = fmt.Fprintf(os.Stderr, "%s (%d:%d)\n", line, tok.Start.Line, tok.Start.Column)
	}
	if traceLexerToken {
		_, _ = fmt.Fprintf(os.Stderr, "Token: %s\n", tok.String())
	}
}

func (l *Lexer) expected(s string, actual rune) {
	l.emitIllegal(fmt.Sprintf("Expected %s but got '%c'(%d)", s, actual, actual))
}

func (l *Lexer) unclosedComment(expected string) {
	l.emitIllegal(fmt.Sprintf("Expected '%s' for closing comment but got EOF", expected))
}

func (l *Lexer) getCurrentPosString(offset int) string {
	var lineEnd int
	for lineEnd = l.current.Offset; lineEnd < len(l.src.Code); lineEnd++ {
		if l.src.Code[lineEnd] == '\n' {
			break
		}
	}
	line := string(l.src.Code[l.current.Offset-l.current.Column+1 : lineEnd])
	line += fmt.Sprintf("\n%"+strconv.Itoa(l.current.Column+offset)+"s", "^")
	return line
}

func (l *Lexer) forward() {
	r, _, err := l.input.ReadRune()
	if err == io.EOF {
		l.top = 0
		l.eof = true
		return
	}

	if err != nil {
		panic(err)
	}

	if !utf8.ValidRune(r) {
		l.emitIllegal(fmt.Sprintf("Invalid UTF-8 character '%c' (%d)", r, r))
		l.eof = true
		return
	}

	l.top = r
	l.eof = false
	if traceLexerPos {
		l.currentPosString = l.getCurrentPosString(0)
		_, _ = fmt.Fprintln(os.Stderr, l.currentPosString)
	}
}

func (l *Lexer) eat() {
	size := utf8.RuneLen(l.top)
	l.current.Offset += size

	// TODO: Consider \n\r
	if l.top == '\n' {
		l.current.Line++
		l.current.Column = 1
	} else {
		l.current.Column += size
	}

	l.forward()
}

func (l *Lexer) consume() {
	if l.eof {
		return
	}
	l.eat()
	l.start = l.current
}

func (l *Lexer) errmsg(msg string) {
	if l.Error == nil {
		return
	}
	l.Error(msg, l.current)
}

func (l *Lexer) eatIdent() bool {
	if !isLetter(l.top) {
		l.expected("letter for head character of identifer", l.top)
		return false
	}
	l.eat()

	for isLetter(l.top) || isDigit(l.top) {
		l.eat()
	}
	return true
}

func lexComment(l *Lexer) stateFn {
	lastRune := l.top
	for {
		if l.eof {
			l.unclosedComment("*")
			return nil
		}
		if l.top == ')' && lastRune == '*' {
			l.eat()
			l.emit(token.COMMENT)
			return lex
		}
		if l.top == '*' {
			l.eat()
			if l.eof {
				l.unclosedComment(")")
				return nil
			}
			if l.top == ')' {
				l.eat()
				l.emit(token.COMMENT)
				return lex
			}
		}
		lastRune = l.top
		l.eat()
	}
}

func lexWhiteSpace(l *Lexer) stateFn {
	l.eat()
	for {
		if l.eof {
			return lex
		}
		if l.top != '\n' && l.top != '\r' && unicode.IsSpace(l.top) {
			l.eat()
		} else {
			break
		}
	}
	l.emit(token.WHITESPACE)
	return lex
}

// e.g. 123.45e10
func lexNumber(l *Lexer) stateFn {
	first := l.top
	// Eat first digit. It's known as digit in lex()
	l.eat()
	if first == '0' && l.top == 'x' {
		l.eat()
		return lexHex
	}
	for isDigit(l.top) {
		l.eat()
	}
	l.emit(token.NUM_LIT)
	return lex
}

func lexHex(l *Lexer) stateFn {
	if !isHex(l.top) {
		l.emitIllegal("illegal hex digit")
	}
	for isHex(l.top) {
		l.eat()
	}
	l.emit(token.HEX_LIT)
	return lex
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z' ||
		r == '_' ||
		r >= utf8.RuneSelf && unicode.IsLetter(r)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isHex(r rune) bool {
	return 'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z' ||
		'0' <= r && r <= '9'
}

func lexIdent(l *Lexer) stateFn {
	if l.top == '_' {
		l.eat()
		if !isLetter(l.top) {
			l.emit(token.SPID)
			return lex
		}
	}
	// ( : - < = already filtered out
	for idx, s := range token.SeparatorTable {
		if s != "" && l.top == rune(s[0]) {
			l.eat()
			l.emit(token.Kind(idx))
			return lex
		}
	}
	// eat a TID
	var isTid bool
	if l.top == '\'' {
		l.eat()
		isTid = true
	}
	if !l.eatIdent() {
		if isTid {
			l.emitIllegal("Expected <Type Parameter Name> after ' but got 'Nothing'")
		}
		return nil
	}
	i := string(l.src.Code[l.start.Offset:l.current.Offset])
	if isTid {
		if 'A' <= i[1] && i[1] <= 'Z' {
			l.emit(token.TID)
		} else {
			l.emitIllegal("Type Parameter Name should be capitalized")
		}
		return lex
	}
	if l.emitPrimeType(i) {
		return lex
	}
	l.emitIdent(i)
	return lex
}

func lexStringLiteral(l *Lexer) stateFn {
	l.eat() // Eat first '"'
	for !l.eof {
		if l.top == '\\' {
			// Skip escape ('\' and next char)
			l.eat()
			l.eat()
		}
		if l.top == '"' {
			l.eat()
			l.emit(token.STRING_LIT)
			return lex
		}
		l.eat()
	}
	l.emitIllegal("Unclosed string literal")
	return nil
}

func lex(l *Lexer) stateFn {
	for {
		if l.eof {
			l.emit(token.EOF)
			return nil
		}

		// switch multi-sep and literals
		switch l.top {
		case '(': // ( or (*
			l.eat()
			if l.top == '*' {
				l.eat()
				return lexComment
			}
			l.emit(token.LPAREN)
		case ':': // : or :=
			l.eat()
			if l.top == '=' {
				l.eat()
				l.emit(token.ASSIGN)
			} else {
				l.emit(token.COLON)
			}
		case '-': // ->
			l.eat()
			if l.top == '>' {
				l.eat()
				l.emit(token.TARROW)
			} else if isDigit(l.top) {
				for isDigit(l.top) {
					l.eat()
				}
				l.emit(token.NUM_LIT)
			} else {
				l.emitIllegal("Illegal operator '-'")
			}
		case '<': // <-
			l.eat()
			if l.top == '-' {
				l.eat()
				l.emit(token.FETCH)
			} else {
				l.emitIllegal("Illegal operator '<'")
			}
		case '=': // = or =>
			l.eat()
			if l.top == '>' {
				l.eat()
				l.emit(token.ARROW)
			}
			l.emit(token.EQ)
		case '"': // read string
			return lexStringLiteral
		case '\n': // Unix
			l.eat()
			l.emit(token.NEWLINE)
		case '\r': // Old Mac
			l.eat()
			if l.top == '\n' { // Windows
				l.eat()
			}
			l.emit(token.NEWLINE)
		default:
			switch {
			case unicode.IsSpace(l.top):
				return lexWhiteSpace
			case isDigit(l.top): // int literal or hex literal
				return lexNumber
			default:
				return lexIdent
			}
		}
	}
}
