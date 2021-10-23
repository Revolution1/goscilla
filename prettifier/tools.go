package prettifier

import (
	"bytes"
	"container/list"
	"github.com/sirupsen/logrus"
	"goscilla/token"
)

var spaceTok = token.NewOrphanToken(token.WHITESPACE, " ")

func isSpace(k token.Kind) bool {
	return k == token.WHITESPACE
}

func isAllSpace(k token.Kind) bool {
	return k == token.NEWLINE || isSpace(k)
}

func isPureCommentTokLine(line []*token.Token) bool {
	for _, t := range line {
		if !isAllSpace(t.Kind) && t.Kind != token.COMMENT {
			return false
		}
	}
	return true
}

func isTopOfFile(e *list.Element) bool {
	return isNilElm(prevNonSpaceElm(e))
}

func trimSpaceLeft(p *prettifier, e *list.Element) {
	i := e.Prev()
	if isNilElm(i) || i.Value.(*token.Token).Kind == token.NEWLINE || !isSpace(i.Value.(*token.Token).Kind) {
		return
	}
	p.tokenList.Remove(i)
	trimSpaceLeft(p, e)
}

func trimSpaceLeftDry(e *list.Element) *list.Element {
	i := e.Prev()
	if isNilElm(i) || i.Value.(*token.Token).Kind == token.NEWLINE || !isSpace(i.Value.(*token.Token).Kind) {
		return i
	}
	return trimSpaceLeftDry(e)
}

func trimAllSpaceLeft(p *prettifier, e *list.Element) {
	logrus.Traceln(getAllString(p.tokenList.Front()))
	i := e.Prev()
	if isNilElm(i) || !isAllSpace(i.Value.(*token.Token).Kind) {
		return
	}
	p.tokenList.Remove(i)
	trimAllSpaceLeft(p, e)
}

func firstNonSpaceTokenOfLine(e *list.Element) bool {
	i := e.Prev()
	if isNilElm(i) {
		return true
	}
	t := i.Value.(*token.Token)
	if t.Kind == token.NEWLINE {
		return true
	}
	if !isSpace(t.Kind) {
		return false
	}
	return firstNonSpaceTokenOfLine(i)
}

func lastElmOfPrevLine(e *list.Element) *list.Element {
	i := e.Prev()
	for !isNilElm(i) { // goto prev line
		t := i.Value.(*token.Token)
		if t.Kind == token.NEWLINE {
			break
		}
		i = i.Prev()
	}
	return prevNonSpaceElm(i)
}

func prevNonSpaceElm(e *list.Element) *list.Element {
	i := e.Prev()
	if isNilElm(i) {
		return i
	}
	if !isAllSpace(i.Value.(*token.Token).Kind) {
		return i
	}
	return prevNonSpaceElm(i)
}

func nextNonSpaceElm(e *list.Element) *list.Element {
	i := e.Next()
	for !isNilElm(i) {
		t := i.Value.(*token.Token)
		if !isAllSpace(t.Kind) {
			return i
		}
		i = i.Next()
	}
	return nil
}

func isByStr20(t *token.Token) bool {
	return t.Value() == "ByStr20" && t.Kind == token.BYSTR_TYPE
}

func trimOneIndent(p *prettifier, e *list.Element) {
	if !isNilElm(e.Prev()) && e.Prev().Value.(*token.Token).Kind == token.WHITESPACE {
		p.tokenList.Remove(e.Prev())
	}
}

func lastElmOfKindBefore(k token.Kind, e *list.Element) *list.Element {
	i := e.Prev()
	for !isNilElm(i) {
		t := i.Value.(*token.Token)
		if t.Kind == k {
			return i
		}
		i = i.Prev()
	}
	return nil
}

func firstElmOfKindAfter(k token.Kind, e *list.Element) *list.Element {
	i := e.Next()
	for !isNilElm(i) {
		t := i.Value.(*token.Token)
		if t.Kind == k {
			return i
		}
		i = i.Next()
	}
	return nil
}

func isNilElm(e *list.Element) bool {
	return e == nil || e.Value == nil
}

func lineOfElm(e *list.Element) (*list.Element, *list.Element) {
	start, end := e, e
	for !isNilElm(start.Prev()) {
		t := start.Prev().Value.(*token.Token)
		if t.Kind == token.NEWLINE {
			break
		}
		start = start.Prev()
	}
	if end.Value.(*token.Token).Kind != token.NEWLINE {
		for !isNilElm(end.Next()) {
			t := end.Next().Value.(*token.Token)
			if t.Kind == token.NEWLINE {
				break
			}
			end = end.Next()
		}
	}
	return start, end
}

func getLineString(e *list.Element) string {
	var line string
	i, _ := lineOfElm(e)
	for !isNilElm(i.Next()) {
		t := i.Next().Value.(*token.Token)
		if t.Kind == token.NEWLINE {
			break
		}
		line = line + i.Value.(*token.Token).Value()
		i = i.Next()
	}
	return line
}

func getAllString(e *list.Element) string {
	if !isNilElm(e) {
		for !isNilElm(e.Prev()) {
			e = e.Prev()
		}
	}
	buf := &bytes.Buffer{}
	for ; e != nil; e = e.Next() {
		buf.WriteString(e.Value.(*token.Token).Value())
	}
	return buf.String()
}

func getIndent(str string, n int) (indent string) {
	for n > 0 {
		indent = indent + str
		n--
	}
	return
}

func insertIndents(p *prettifier, e *list.Element) {
	logrus.Traceln("try indent on:", e.Value.(*token.Token).Value())
	if firstNonSpaceTokenOfLine(e) {
		logrus.Traceln("before  trim:", getLineString(e))
		trimSpaceLeft(p, e)
		logrus.Traceln("after   trim:", getLineString(e))
		for _, _ = range p.contexts {
			p.tokenList.InsertBefore(token.NewOrphanToken(token.WHITESPACE, p.option.IndentStr), e)
		}
		logrus.Traceln("after indent:", getLineString(e))
	}
}
