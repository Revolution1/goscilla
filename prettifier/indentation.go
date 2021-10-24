package prettifier

import (
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
			if !isNilElm(e.Prev()) && len(p.contexts) > 0 {
				prev := e.Prev().Value.(*token.Token)
				switch p.contexts[0] {
				case token.LET, token.TRANSITION, token.PROCEDURE:
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
			// keywords always on top of line
			case token.IMPORT, token.LIBRARY, token.SCILLA_VERSION, token.TRANSITION, token.PROCEDURE, token.TYPE:
				p.clearContext()
				trimSpaceLeft(p, e)
				switch t.Kind {
				case token.TRANSITION, token.PROCEDURE, token.TYPE:
					p.pushContext(t.Kind)
				}
			case token.FIELD:
				p.pushContext(t.Kind)
				insertIndents(p, e)
				//if len(p.contexts) > 0 {
				//	insertIndents(p, e)
				//} else {
				//	trimSpaceLeft(p, e)
				//}
			case token.WITH:
				insertIndents(p, e)
				prev := prevNonSpaceElm(e)
				if !isNilElm(prev) && isByStr20(prev.Value.(*token.Token)) {
					trimAllSpaceLeft(p, e) // ByStr20 with always on one line
					p.tokenList.InsertBefore(spaceTok, e)
					p.pushContext(token.BYSTR_TYPE)
				} else if len(p.contexts) > 0 {
					if p.tailContext(0) != token.MATCH {
						p.pushContext(token.WITH)
					}
				} else {
					p.pushContext(token.WITH)
				}
			case token.MATCH:
				if !firstNonSpaceTokenOfLine(e) {
					p.tokenList.InsertBefore(newLineTok, e)
				}
				insertIndents(p, e)
				p.pushContext(t.Kind)
			case token.FUN, token.TFUN:
				if !firstNonSpaceTokenOfLine(e) {
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
				if !firstNonSpaceTokenOfLine(e) { // need a newline before it
					p.tokenList.InsertBefore(newLineTok, e)
				}
				if len(p.contexts) > 0 { // upgrade to top tier
					var upgrade bool
					if p.contexts[0] == token.TYPE {
						upgrade = true
					} else if p.contexts[0] == token.LET {
						lastElm := lastElmOfPrevLine(e)
						if lastElm != nil {
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
						trimSpaceLeft(p, e)
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
				if !firstNonSpaceTokenOfLine(e) {
					p.tokenList.InsertBefore(newLineTok, e)
				}
				if len(p.contexts) > 0 && p.tailContext(0) == t.Kind {
					p.popContext()
				}
				insertIndents(p, e)
				if !p.option.IndentBar {
					trimOneIndent(p, e)
				}
				p.pushContext(t.Kind)
			case token.ARROW:
				if len(p.contexts) > 0 && p.tailContext(0) == token.WITH {
					p.popContext()
				}
				insertIndents(p, e)
			case token.END:
			Loop:
				for len(p.contexts) > 0 {
					switch p.tailContext(0) {
					case token.MATCH, token.BYSTR_TYPE, token.TRANSITION, token.PROCEDURE:
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
			case token.LBRACE, token.LPAREN, token.LSQB:
				insertIndents(p, e)
				p.pushContext(t.Kind)
			case token.RBRACE, token.RPAREN, token.RSQB:
				insertIndents(p, e)
				if len(p.contexts) > 0 && p.tailContext(0) == t.Kind-1 {
					p.popContext()
				}
				trimOneIndent(p, e)
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
