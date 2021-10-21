package prettifier

import (
	"container/list"
	"github.com/sirupsen/logrus"
	"goscilla/token"
)

func indent(p *prettifier) error {
	var (
		aroundComments bool
		//indentTok      = token.NewOrphanToken(token.INDENT, p.option.IndentStr)
		newLineTok = token.NewOrphanToken(token.NEWLINE, p.option.NewLineStr)
	)
	// first scan:
	e := p.tokenList.Front()
	for e != nil {
		t := e.Value.(*token.Token)
		switch t.Kind {
		case token.WHITESPACE:
		case token.NEWLINE:
			if aroundComments {
				e = e.Next()
				continue
			}
			// try pop context based on token at the end of line
			if e.Prev() != nil && e.Prev().Value != nil && len(p.contexts) > 0 {
				prev := e.Prev().Value.(*token.Token)
				switch p.contexts[0] {
				case token.LET:
					switch prev.Kind {
					// name and lit pop all let or type context until hit a bar
					case token.ID, token.CID, token.SPID,
						token.NUM_LIT, token.STRING_LIT, token.TRUE, token.FALSE, token.ZERO,
						token.RBRACE:
						for len(p.contexts) > 0 && p.tailContext(0) != token.BAR { // pop other context
							logrus.Tracef("loop pop context %s for let ends by %s", token.TokenTable[p.tailContext(0)], prev)
							p.popContext()
						}
						if len(p.contexts) > 0 {
							logrus.Tracef("pop context %s for let ends by %s", token.TokenTable[p.tailContext(0)], prev)
							p.popContext() // pop bar context
						}
					}
				case token.TYPE:
					switch prev.Kind {
					// end a bar of type context
					case token.INT_TYPE, token.STRING_TYPE, token.BYSTR_TYPE, token.BNUM_TYPE, token.MESSAGE_TYPE, token.EVENT_TYPE,
						token.BOOL, token.NAT, token.OPTION, token.LIST,
						token.ID, token.CID:
						for len(p.contexts) > 0 {
							if p.tailContext(0) == token.BAR {
								logrus.Tracef("pop context %s to bar for type ends by %s", token.TokenTable[p.tailContext(0)], prev)
								p.popContext()
								break
							}
							logrus.Tracef("pop context %s for type ends by %s", token.TokenTable[p.tailContext(0)], prev)
							p.popContext()
						}
					}
				}
			}
		case token.COMMENT:
			aroundComments = true
		default:
			aroundComments = false
			switch t.Kind {
			case token.IMPORT, token.LIBRARY, token.SCILLA_VERSION, token.CONTRACT, token.TRANSITION, token.PROCEDURE, token.TYPE, token.FIELD:
				p.clearContext()
				switch t.Kind {
				case token.TRANSITION, token.PROCEDURE, token.TYPE:
					p.pushContext(t.Kind)
				}
			case token.MATCH:
				if !firstTokenOfLine(e) {
					p.tokenList.InsertBefore(newLineTok, e)
				}
				insertIndents(p, e)
				p.pushContext(t.Kind)
			case token.FUN, token.TFUN:
				if !firstTokenOfLine(e) {
					p.tokenList.InsertBefore(newLineTok, e)
				}
				insertIndents(p, e)
				if len(p.contexts) == 0 {
					p.pushContext(t.Kind)
				} else if p.tailContext(0) != token.FUN && p.tailContext(0) != token.TFUN {
					p.pushContext(t.Kind)
				} else {
					trimOneIndent(p, e)
				}
			case token.LET:
				if !firstTokenOfLine(e) { // need a newline before it
					p.tokenList.InsertBefore(newLineTok, e)
				}
				if len(p.contexts) > 0 { // upgrade to top tier
					var upgrade bool
					if p.contexts[0] == token.TYPE{
						upgrade = true
					} else if p.contexts[0] == token.LET {
						lastElm := lastElmOfPrevLine(e)
						if lastElm != nil && lastElm.Value != nil {
							tok := lastElm.Value.(*token.Token)
							switch tok.Kind {
							case token.IN, token.TARROW, token.ARROW, token.FETCH, token.SEMICOLON, token.EQ:
							default:
								upgrade = true
							}
						}
					}
					if upgrade {
						p.clearContext()
						trimLeft(p, e)
					}
				}
				insertIndents(p, e)
				if len(p.contexts) == 0 {
					p.pushContext(t.Kind)
				} else if p.tailContext(0) != token.LET {
					p.pushContext(t.Kind)
				} else {
					trimOneIndent(p, e)
				}
			case token.BAR:
				if !firstTokenOfLine(e) {
					p.tokenList.InsertBefore(newLineTok, e)
				}
				insertIndents(p, e)
				if !p.option.IndentBar {
					trimOneIndent(p, e)
				}
				p.pushContext(t.Kind)
			case token.END:
			Loop:
				for len(p.contexts) > 0 {
					switch p.tailContext(0) {
					case token.MATCH, token.WITH, token.TRANSITION, token.PROCEDURE:
						p.popContext() // pop above context
						break Loop
					default:
						p.popContext() // pop other context
					}
				}
				insertIndents(p, e)
			case token.IN:
				insertIndents(p, e)
				if p.tailContext(0) == token.LET {
					trimOneIndent(p, e)
				}
			default:
				insertIndents(p, e)
			}
		}
		if e != nil {
			e = e.Next()
		}
	}
	return nil
}

func insertIndents(p *prettifier, e *list.Element) {
	if firstTokenOfLine(e) {
		trimLeft(p, e)
		for _, _ = range p.contexts {
			p.tokenList.InsertBefore(token.NewOrphanToken(token.INDENT, p.option.IndentStr), e)
		}
	}
}

func trimLeft(p *prettifier, e *list.Element) {
	i := e.Prev()
	for i != nil && i.Value != nil {
		tok := i.Value.(*token.Token)
		if tok.Kind == token.WHITESPACE {
			p.tokenList.Remove(i)
			i = e.Prev()
			continue
		}
		break
	}
}

func firstTokenOfLine(e *list.Element) bool {
	i := e.Prev()
	for i != nil && i.Value != nil {
		t := i.Value.(*token.Token)
		if t.Kind == token.WHITESPACE {
			i = i.Prev()
		} else if t.Kind == token.NEWLINE {
			return true
		} else {
			return false
		}
	}
	return true
}

func lastElmOfPrevLine(e *list.Element) *list.Element {
	i := e.Prev()
	for i != nil && i.Value != nil {
		t := i.Value.(*token.Token)
		if t.Kind == token.WHITESPACE || t.Kind == token.NEWLINE {
			i = i.Prev()
		} else {
			return i
		}
	}
	return i
}

func trimOneIndent(p *prettifier, e *list.Element) {
	if e.Prev() != nil && e.Prev().Value != nil && e.Prev().Value.(*token.Token).Kind == token.INDENT {
		p.tokenList.Remove(e.Prev())
	}
}
