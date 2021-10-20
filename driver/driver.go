// Package driver is amediator to glue all packages for GoCaqml. provides a compiler function for GoCaml codes.
// It provides compiler functinalities for GoCaml.
package driver

import (
	"fmt"
	"github.com/rhysd/locerr"
	"github.com/sirupsen/logrus"
	"goscilla/syntax"
	"goscilla/token"
	"os"
)

type OptLevel int

const (
	O0 OptLevel = iota
	O1
	O2
	O3
)

// Driver instance to compile GoCaml code into other representations.
type Driver struct {
}

// Lex PrintTokens returns the lexed tokens for a source code.
func (d *Driver) Lex(src *locerr.Source) chan token.Token {
	l := syntax.NewLexer(src)
	l.Error = func(msg string, pos locerr.Pos) {
		err := locerr.ErrorAt(pos, msg)
		err.PrintToFile(os.Stderr)
		_, _ = fmt.Fprintln(os.Stderr)
	}
	go l.Lex()
	return l.Tokens
}

// PrintTokens show list of tokens lexed.
func (d *Driver) PrintTokens(src *locerr.Source) {
	tokens := d.Lex(src)
	for t := range tokens {
		fmt.Println(t.String())
		switch t.Kind {
		case token.EOF, token.ILLEGAL:
			return
		}
	}
}

// Prettify print prettified code.
func (d *Driver) Prettify(src *locerr.Source) {
	tokens := d.Lex(src)
	err := syntax.Prettify(tokens, os.Stdout, nil)
	if err != nil {
		logrus.Error(err)
	}
}

// Parse parses the source and returns the parsed AST.
//func (d *Driver) Parse(src *locerr.Source) (*ast.AST, error) {
//return syntax.Parse(src)
//}

// PrintAST outputs AST structure to stdout.
//func (d *Driver) PrintAST(src *locerr.Source) {
//	a, err := d.Parse(src)
//	if err != nil {
//		_, _ = fmt.Fprintln(os.Stderr, err.Error())
//		return
//	}
//	ast.Println(a)
//}
