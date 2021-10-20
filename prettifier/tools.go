package prettifier

import (
	"container/list"
	"goscilla/token"
)

func isWhiteSpace(t displayToken) bool {
	if t.kind == token.WHITESPACE || t.kind == token.NEWLINE {
		return true
	}
	return false
}

func isAllWhiteSpace(line []displayToken) bool {
	for _, t := range line {
		if t.kind != token.WHITESPACE && t.kind != token.NEWLINE {
			return false
		}
	}
	return true
}

func isPureCommentLine(line []displayToken) bool {
	for _, t := range line {
		if t.kind != token.WHITESPACE && t.kind != token.NEWLINE && t.kind != token.COMMENT {
			return false
		}
	}
	return true
}

func isPureCommentTokLine(line []*token.Token) bool {
	for _, t := range line {
		if t.Kind != token.WHITESPACE && t.Kind != token.NEWLINE && t.Kind != token.COMMENT {
			return false
		}
	}
	return true
}

var spaceTok = token.NewOrphanToken(token.WHITESPACE, " ")

func isTopOfFile(e *list.Element) bool {
	var insertPoint *list.Element
	for insertPoint = e.Prev(); insertPoint != nil && insertPoint.Value != nil; insertPoint = insertPoint.Prev() {
		it := insertPoint.Value.(*token.Token)
		if it.Kind == token.WHITESPACE || it.Kind == token.NEWLINE {
			continue
		}
		break
	}
	if insertPoint == nil || insertPoint.Value == nil { // it's the top non-space line of source file
		return true
	}
	return false
}

func prevLine(e *list.Element) ([]*token.Token, *list.Element) {
	var line []*token.Token
	if e != nil && e.Value != nil && e.Value.(*token.Token).Kind != token.NEWLINE {
		for e != nil {
			e = e.Prev()
			if e != nil && e.Value != nil && e.Value.(*token.Token).Kind == token.NEWLINE {
				break
			}
		}
	}
	e = e.Prev()
	if e == nil || e.Value == nil || e.Value.(*token.Token).Kind == token.NEWLINE {
		return line, e
	}
	for e != nil {
		if e != nil && e.Value != nil && e.Value.(*token.Token).Kind != token.NEWLINE {
			line = append(line, e.Value.(*token.Token))
			e = e.Prev()
			continue
		}
		break
	}
	return line, e
}

func nextLine(e *list.Element) ([]*token.Token, *list.Element) {
	return nil, e
}
