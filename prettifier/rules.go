package prettifier

import (
	"container/list"
	"goscilla/token"
)

// rules available rule list, ps: order matters
var rules = []Rule{
	noEmptyLineAtStartOfFile,
	onlyOneEmptyLine,
	noOnlySpaceLine,
	allMidLineSpaceSingle,

	indent,

	spaceBeforeTokens,
	spaceAfterTokens,
	noSpaceBeforeTokens,
	noSpaceAfterTokens,

	noTrailingSpace,
	cleanLinesBeforeEOF,
}

func doNothing(p *prettifier) error {
	e := p.tokenList.Front()
	for e != nil {
		t := e.Value.(*token.Token)
		// actual code
		_ = t
		// actual code
		e = e.Next()
	}
	return nil
}

func noTrailingSpace(p *prettifier) error {
	for e := p.tokenList.Front(); e != nil; e = e.Next() {
		t := e.Value.(*token.Token)
		if t.Kind == token.NEWLINE {
			for e.Prev() != nil && e.Prev().Value != nil {
				prevKind := e.Prev().Value.(*token.Token).Kind
				if prevKind == token.WHITESPACE || prevKind == token.INDENT {
					p.tokenList.Remove(e.Prev())
				} else {
					break
				}
			}
		}
	}
	return nil
}

func cleanLinesBeforeEOF(p *prettifier) error {
	e := p.tokenList.Back()
	if e != nil {
		e = e.Prev()
	}
	for e != nil && e.Value != nil {
		t := e.Value.(*token.Token)
		if t.Kind == token.NEWLINE || t.Kind == token.WHITESPACE {
			e = e.Prev()
			p.tokenList.Remove(e.Next())
			continue
		}
		break
	}
	if p.option.EndWithNewLine {
		p.tokenList.InsertBefore(token.NewOrphanToken(token.NEWLINE, p.option.NewLineStr), p.tokenList.Back())
	}
	return nil
}

func noEmptyLineAtStartOfFile(p *prettifier) error {
	e := p.tokenList.Front()
	for e != nil {
		t := e.Value.(*token.Token)
		if t.Kind == token.NEWLINE || t.Kind == token.WHITESPACE {
			prev := e
			e = e.Next()
			p.tokenList.Remove(prev)
			continue
		} else {
			break
		}
	}
	return nil
}

func allMidLineSpaceSingle(p *prettifier) error {
	for e := p.tokenList.Front(); e != nil; e = e.Next() {
		t := e.Value.(*token.Token)
		if t.Kind == token.WHITESPACE && t.Value() != " " {
			if e.Prev() != nil && e.Prev().Value != nil {
				prevKind := e.Prev().Value.(*token.Token).Kind
				if prevKind != token.NEWLINE && prevKind != token.COMMENT {
					t.SetLiteral(" ")
				}
			}
		}
	}
	return nil
}

func noSpaceBeforeTokens(p *prettifier) error {
	for e := p.tokenList.Front(); e != nil; e = e.Next() {
		t := e.Value.(*token.Token)
		switch t.Kind {
		case token.COLON, token.COMMA, token.SEMICOLON, token.RSQB, token.RPAREN, token.RBRACE, token.PERIOD:
			if e.Prev() != nil && e.Prev().Value != nil && e.Prev().Value.(*token.Token).Kind == token.WHITESPACE {
				p.tokenList.Remove(e.Prev())
			} else {
				break
			}
		}
	}
	return nil
}

func noSpaceAfterTokens(p *prettifier) error {
	for e := p.tokenList.Front(); e != nil; e = e.Next() {
		t := e.Value.(*token.Token)
		switch t.Kind {
		case token.LPAREN, token.LBRACE, token.LSQB, token.AT, token.AND:
			if e.Next().Value.(*token.Token).Kind == token.WHITESPACE {
				p.tokenList.Remove(e.Next())
			} else {
				break
			}
		}
	}
	return nil
}

func spaceBeforeTokens(p *prettifier) error {
	for e := p.tokenList.Front(); e != nil; e = e.Next() {
		t := e.Value.(*token.Token)
		switch t.Kind {
		case token.EQ, token.TARROW, token.ARROW, token.FETCH:
			if e.Prev() != nil && e.Prev().Value != nil && e.Prev().Value.(*token.Token).Kind != token.WHITESPACE {
				p.tokenList.InsertBefore(spaceTok, e)
			} else {
				break
			}
		}
	}
	return nil
}

func spaceAfterTokens(p *prettifier) error {
	for e := p.tokenList.Front(); e != nil; e = e.Next() {
		t := e.Value.(*token.Token)
		switch t.Kind {
		case token.EQ, token.TARROW, token.ARROW, token.FETCH:
			if e.Next().Value.(*token.Token).Kind != token.WHITESPACE {
				p.tokenList.InsertAfter(spaceTok, e)
			} else {
				break
			}
		}
	}
	return nil
}

func noOnlySpaceLine(p *prettifier) error {
	e := p.tokenList.Front()
	for e != nil && e.Value != nil {
		t := e.Value.(*token.Token)
		if e.Next() == nil {
			break
		}
		nextTok := e.Next().Value.(*token.Token)
		if t.Kind == token.WHITESPACE && (nextTok.Kind == token.NEWLINE || nextTok.Kind == token.EOF) {
			e = e.Next()
			p.tokenList.Remove(e.Prev())
		}
		e = e.Next()
	}
	return nil
}

func onlyOneEmptyLine(p *prettifier) error {
	e := p.tokenList.Front()
	for e != nil {
		if e.Value.(*token.Token).Kind == token.NEWLINE { // the first is a newline at end of non-empty line
			if e.Prev() != nil && e.Prev().Value != nil {
				e = e.Next()
			}
			if e != nil && e.Value.(*token.Token).Kind == token.WHITESPACE { // skip space in case of only-space line
				e = e.Next()
			}
			if e == nil || e.Value.(*token.Token).Kind != token.NEWLINE { // the empty line that allowed
				continue
			}
			// remove all empty new line from here till non-newline
			var space *list.Element
			e = e.Next()
			for e != nil {
				kind := e.Value.(*token.Token).Kind
				if kind == token.WHITESPACE {
					space = e
					e = e.Next()
				} else if kind == token.NEWLINE || kind == token.EOF {
					if space != nil {
						p.tokenList.Remove(space)
					}
					e = e.Next()
					if e != nil {
						p.tokenList.Remove(e.Prev())
					}
				} else {
					break
				}
			}
		}
		if e != nil {
			e = e.Next()
		}
	}
	return nil
}
