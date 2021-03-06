// Package token defines tokens of GoScilla source codes.
package token

import (
	"fmt"
	"github.com/rhysd/locerr"
	"strconv"
)

type Kind int

const (
	ILLEGAL Kind = iota

	// from https://github.com/Zilliqa/scilla/blob/master/src/base/ScillaLexer.mll
	// whitespaces
	NEWLINE    // \n \r \r\n
	COMMENT    // (* xxx *)
	WHITESPACE // " " \t

	// Literals
	STRING_LIT // "xxxx"
	NUM_LIT    // 123
	HEX_LIT    // 0x1234abcd

	// Prime types
	INT_TYPE     // IntXX UintXX
	STRING_TYPE  // String
	BYSTR_TYPE   // ByStrXX
	BNUM_TYPE    // BNum
	MESSAGE_TYPE // Message
	EVENT_TYPE   // Event

	// Keywords
	FORALL
	BUILTIN
	LIBRARY
	IMPORT
	LET
	IN
	MATCH
	WITH
	END
	FUN
	TFUN
	CONTRACT
	TRANSITION
	SEND
	EVENT
	FIELD
	ACCEPT
	EXISTS
	DELETE
	EMP
	MAP
	SCILLA_VERSION
	TYPE
	OF
	TRY
	CATCH
	AS
	PROCEDURE
	THROW

	// Separators
	SEMICOLON  // ;
	COLON      // :
	PERIOD     // .
	BAR        // |
	LSQB       // [
	RSQB       // ]
	LPAREN     // (
	RPAREN     // )
	LBRACE     // {
	RBRACE     // }
	COMMA      // ,
	ARROW      // =>
	TARROW     // ->
	EQ         // =
	AND        // &
	FETCH      // <-
	ASSIGN     // :=
	AT         // @
	UNDERSCORE // _

	// Identifiers
	ID   // simple name          [a-z][A-Za-z0-9_]*
	CID  // qualified name       [A-Z][A-Za-z0-9_]*
	TID  // type parameter name '[A-Z][A-Za-z0-9_]*
	SPID // special const name  _[A-Za-z0-9_]*

	// from https://github.com/Zilliqa/scilla/blob/master/src/base/Datatypes.ml
	// Builtin ADT
	BOOL  // Bool
	TRUE  // True
	FALSE // False

	NAT  // Nat
	ZERO // Zero
	SUCC // Succ

	OPTION
	SOME
	NONE

	LIST
	CONS
	NIL

	PAIR

	// only for pretty formatting
	//INDENT
	//ALIGN
	//ADDRESS_WITH

	// Other tokens
	EOF
)

var TokenTable = [EOF + 1]string{
	// whitespaces
	NEWLINE:    "NEWLINE",
	COMMENT:    "COMMENT",
	WHITESPACE: "WHITESPACE",

	// Literals
	STRING_LIT: "STRING_LIT",
	NUM_LIT:    "NUM_LIT",
	HEX_LIT:    "HEX_LIT",

	// Identifiers
	ID:   "ID",
	CID:  "CID",
	TID:  "TID",
	SPID: "SPID",

	// others
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
}

// PrimeTypeTable from https://github.com/Zilliqa/scilla/blob/master/src/base/ScillaParser.mly#L31
var PrimeTypeTable = [...]string{
	// Integer types
	INT_TYPE:     "Int",
	BYSTR_TYPE:   "ByStr",
	BNUM_TYPE:    "BNum",
	MESSAGE_TYPE: "Message",
	EVENT_TYPE:   "EventType",
	STRING_TYPE:  "String",
}

// KeywordTable from https://github.com/Zilliqa/scilla/blob/master/src/base/ScillaLexer.mll#L68
var KeywordTable = [...]string{
	FORALL:         "forall",
	BUILTIN:        "builtin",
	LIBRARY:        "library",
	IMPORT:         "import",
	LET:            "let",
	IN:             "in",
	MATCH:          "match",
	WITH:           "with",
	END:            "end",
	FUN:            "fun",
	TFUN:           "tfun",
	CONTRACT:       "contract",
	TRANSITION:     "transition",
	SEND:           "send",
	EVENT:          "event",
	FIELD:          "field",
	ACCEPT:         "accept",
	EXISTS:         "exists",
	DELETE:         "delete",
	EMP:            "Emp",
	MAP:            "Map",
	SCILLA_VERSION: "scilla_version",
	TYPE:           "type",
	OF:             "of",
	TRY:            "try",
	CATCH:          "catch",
	AS:             "as",
	PROCEDURE:      "procedure",
	THROW:          "throw",
}

// SeparatorTable from https://github.com/Zilliqa/scilla/blob/master/src/base/ScillaLexer.mll#L100
var SeparatorTable = [...]string{
	SEMICOLON:  ";",
	COLON:      ":",
	PERIOD:     ".",
	BAR:        "|",
	LSQB:       "[",
	RSQB:       "]",
	LPAREN:     "(",
	RPAREN:     ")",
	LBRACE:     "{",
	RBRACE:     "}",
	COMMA:      ",",
	ARROW:      "=>",
	TARROW:     "->",
	EQ:         "=",
	AND:        "&",
	FETCH:      "<-",
	ASSIGN:     ":=",
	AT:         "@",
	UNDERSCORE: "_",
}

// BuiltinADTTable from https://github.com/Zilliqa/scilla/blob/master/src/base/Datatypes.ml
var BuiltinADTTable = [...]string{
	BOOL:  "Bool",
	TRUE:  "Ture",
	FALSE: "False",

	NAT:  "Nat",
	ZERO: "Zero",
	SUCC: "Succ",

	OPTION: "Option",
	SOME:   "Some",
	NONE:   "None",

	LIST: "List",
	CONS: "Cons",
	NIL:  "Nil",

	PAIR: "Pair",
}

func init() {
	for i, v := range PrimeTypeTable {
		if v != "" {
			TokenTable[i] = v
		}
	}
	for i, v := range KeywordTable {
		if v != "" {
			TokenTable[i] = v
		}
	}
	for i, v := range SeparatorTable {
		if v != "" {
			TokenTable[i] = v
		}
	}
	for i, v := range BuiltinADTTable {
		if v != "" {
			TokenTable[i] = v
		}
	}
}

// Token instance for GoScilla.
// It contains its location information and kind.
type Token struct {
	Kind  Kind
	Start locerr.Pos
	End   locerr.Pos
	File  *locerr.Source
	lit   string
}

func NewOrphanToken(kind Kind, lit string) *Token {
	return &Token{
		Kind:  kind,
		Start: locerr.Pos{},
		End:   locerr.Pos{},
		File:  nil,
		lit:   lit,
	}
}

// String returns an information of token. This method is used mainly for
// debug purpose.
func (tok *Token) String() string {
	return fmt.Sprintf(
		"<%s:%s>(%d:%d:%d-%d:%d:%d)",
		TokenTable[tok.Kind],
		tok.DisplayValue(),
		tok.Start.Line, tok.Start.Column, tok.Start.Offset,
		tok.End.Line, tok.End.Column, tok.End.Offset)
}

// Value returns the corresponding a string part of code.
func (tok *Token) Value() string {
	if tok.lit != "" {
		return tok.lit
	}
	val := string(tok.File.Code[tok.Start.Offset:tok.End.Offset])
	if val == "" && tok.Kind != EOF {
		panic("empty token")
	}
	return val
}

func (tok *Token) SetLiteral(s string) {
	tok.lit = s
}

// DisplayValue returns the corresponding a string part of code with characters escaped.
func (tok *Token) DisplayValue() string {
	return strconv.QuoteToASCII(tok.Value())
}
