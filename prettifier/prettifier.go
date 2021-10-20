package prettifier

import (
	"bytes"
	"container/list"
	"fmt"
	"goscilla/token"
	"io"
	"strings"
	"unicode"
)

type Rule func(*prettifier) error

type PrettyOption struct {
	IndentStr      string
	IndentBar      bool
	MaxLineLength  int
	NewLineStr     string
	EndWithNewLine bool
}

var DefaultPrettyOption = &PrettyOption{
	IndentStr:      "  ",
	MaxLineLength:  120,
	NewLineStr:     "\n",
	EndWithNewLine: true,
	IndentBar:      false,
}

type displayToken struct {
	kind    token.Kind
	display string
}



type prettifier struct {
	tokens    []token.Token
	tokenList list.List
	contexts  []token.Kind
	newLines  int
	lastLine  string
	line      []displayToken
	buf       *bytes.Buffer
	option    PrettyOption
}

func newPrettier(tokens []token.Token, option *PrettyOption) *prettifier {
	return &prettifier{
		tokens: tokens,
		option: *option,
		buf:    bytes.NewBuffer([]byte{}),
	}
}

func (p *prettifier) clearLine() {
	p.line = []displayToken{}
}

func (p *prettifier) tailLine(offset int) displayToken {
	return p.line[len(p.line)-1+offset]
}

func (p *prettifier) pushLine(t displayToken) {
	p.line = append(p.line, t)
}

func (p *prettifier) popLine() displayToken {
	t := p.line[len(p.line)-1]
	p.line = p.line[:len(p.line)-2]
	return t
}

func (p *prettifier) clearContext() {
	p.contexts = p.contexts[:0]
}

func (p *prettifier) tailContext(offset int) token.Kind {
	return p.contexts[len(p.contexts)-1+offset]
}

func (p *prettifier) pushContext(k token.Kind) {
	p.contexts = append(p.contexts, k)
}

func (p *prettifier) popContext() token.Kind {
	if len(p.contexts) == 0 {
		panic("pop a empty context stack")
	}
	k := p.contexts[len(p.contexts)-1]
	p.contexts = p.contexts[:len(p.contexts)-1]
	return k
}

func (p *prettifier) getLine() string {
	var line string
	for _, tok := range p.line {
		line = line + tok.display
	}
	line = strings.TrimRightFunc(line, unicode.IsSpace)
	indents := len(p.contexts) - 1
	//for indents > 0 && len(p.contexts) > 0 && p.contexts[indents-1] == p.line[0].kind {
	//	indents--
	//}
	switch p.line[0].kind {
	case token.LET:
		if strings.HasSuffix(p.lastLine, "in") {
			indents--
		}
	case token.BAR:
		if !p.option.IndentBar {
			indents--
		}
	case token.FUN, token.TFUN:
		last := strings.TrimSpace(p.lastLine)
		if strings.HasPrefix(last, "fun") || strings.HasPrefix(last, "tfun") {
			indents--
		}
	}
	if indents < 0 {
		panic("indents < 0")
	}
	if !isPureCommentLine(p.line) { // only indent non pure comment line
		line = getIndent(p.option.IndentStr, indents) + line
	}
	return line
}

func (p *prettifier) postProcessContext() {
	if len(p.contexts) == 0 {
		return
	}
	if isPureCommentLine(p.line) { // leave pure comment as it is
		return
	}

	switch p.contexts[0] {
	case token.END:
		p.popContext()
	case token.LET:
		switch p.tailLine(0).kind {
		// name and lit pop all let or type context until hit a bar
		case token.ID, token.CID, token.SPID,
			token.NUM_LIT, token.STRING_LIT, token.TRUE, token.FALSE, token.ZERO:
			if len(p.contexts) > 0 && p.tailContext(0) != token.BAR {
				p.popContext()
			}
			p.popContext()
		}
	case token.TYPE:
		switch p.tailLine(0).kind {
		// end a bar of type context
		case token.INT_TYPE, token.STRING_TYPE, token.BYSTR_TYPE, token.BNUM_TYPE, token.MESSAGE_TYPE, token.EVENT_TYPE,
			token.BOOL, token.NAT, token.OPTION, token.LIST,
			token.ID, token.CID:
			for len(p.contexts) > 0 {
				if p.tailContext(0) == token.BAR {
					p.popContext()
					break
				}
				p.popContext()
			}
		}
	}
}

func (p *prettifier) breakLine() error {
	return p.writeLine()
}

func (p *prettifier) writeLine() error {
	if isAllWhiteSpace(p.line) {
		p.clearLine()
		if p.lastLine != "" {
			p.newLines++
			_, err := p.buf.WriteString(p.option.NewLineStr)
			if err != nil {
				return err
			}
		}
		return nil
	}
	line := p.getLine()
	//if len(line) > p.option.MaxLineLength && p.tailLine(0).kind != token.COMMENT{
	//	return p.breakLine()
	//}
	_, err := p.buf.WriteString(line + p.option.NewLineStr)
	if err != nil {
		return err
	}
	if p.newLines > 1 {
		_, err := p.buf.WriteString(p.option.NewLineStr)
		if err != nil {
			return err
		}
	}
	p.newLines = 0
	p.lastLine = line
	p.postProcessContext()
	p.clearLine()
	return nil
}

func (p *prettifier) input2(t token.Token) error {
	switch t.Kind {
	case token.EOF:
		err := p.writeLine()
		if err != nil {
			return err
		}
		if p.lastLine != p.option.NewLineStr && p.option.EndWithNewLine {
			_, err := p.buf.WriteString(p.option.NewLineStr)
			if err != nil {
				return err
			}
		}
		return nil
	case token.NEWLINE:
		//if !isAllWhiteSpace(p.line) {
		//	return p.writeLine()
		//}
		if len(p.line) > 0 && p.tailLine(0).kind == token.COMMENT { // newline between comments
			p.newLines = 0
			err := p.writeLine()
			if err != nil {
				return err
			}
			return nil
		}
		if p.lastLine != "" {
			p.newLines++
		}
		return nil
	case token.COMMENT:
		if p.newLines > 0 {
			err := p.writeLine()
			if err != nil {
				return err
			}
		}
	// only keep whitespaces between comments
	case token.WHITESPACE:
		if len(p.line) > 0 && p.tailLine(0).kind == token.COMMENT {
			p.pushLine(displayToken{token.WHITESPACE, t.Value()})
		}
		return nil
	// top tier keywords
	case token.IMPORT, token.LIBRARY, token.SCILLA_VERSION, token.CONTRACT, token.TRANSITION, token.PROCEDURE, token.TYPE, token.FIELD:
		if isAllWhiteSpace(p.line) {
			//log.Warningf("%s should start at top of line", t.Value())
			p.clearLine()
		}
		if len(p.line) > 0 {
			err := p.writeLine()
			if err != nil {
				return err
			}
		}
		p.clearContext()
		switch t.Kind {
		case token.TRANSITION, token.PROCEDURE, token.TYPE:
			p.pushContext(t.Kind)
		}
	// tokens always stars a new line
	case token.LET, token.FUN, token.TFUN, token.MATCH:
		if len(p.line) > 0 {
			err := p.writeLine()
			if err != nil {
				return err
			}
		}
		if t.Kind == token.LET && len(p.contexts) > 0 { // try upgrade let to top tier
			if p.contexts[0] != token.LET && !strings.HasSuffix(p.lastLine, "in") {
				p.clearContext()
			}
		}
		p.pushContext(t.Kind)
	case token.BAR:
		if len(p.line) > 0 {
			err := p.writeLine()
			if err != nil {
				return err
			}
		}
		p.pushContext(token.BAR)
	case token.END: // end pops a match, with, transition, procedure context
		err := p.writeLine()
		if err != nil {
			return err
		}
	Loop:
		for {
			switch p.tailContext(0) {
			case token.MATCH, token.WITH, token.TRANSITION, token.PROCEDURE:
				break Loop
			default:
				p.popContext()
			}
		}
	// do not add space before these
	case token.COLON, token.COMMA, token.SEMICOLON, token.RSQB, token.RPAREN, token.RBRACE, token.PERIOD:
	default:
		if len(p.line) > 0 {
			switch p.tailLine(0).kind {
			// do not add space after these
			case token.LPAREN, token.LBRACE, token.LSQB, token.AT, token.AND:
			default:
				p.pushLine(displayToken{token.WHITESPACE, " "})
			}
		} else {
			p.pushLine(displayToken{token.WHITESPACE, " "})
		}
	}
	p.pushLine(displayToken{t.Kind, t.Value()})
	// this after pushing current
	switch t.Kind {
	case token.IN:
		if len(p.contexts) > 1 {
			err := p.writeLine()
			if err != nil {
				return err
			}
			p.newLines--
		}
	}
	return nil
}

func (p *prettifier) input(t *token.Token) {
	p.tokenList.PushBack(t)
}

func (p *prettifier) reset() {
	p.line = p.line[:0]
	p.contexts = p.contexts[:0]
	p.newLines = 0
	p.lastLine = ""
}

func (p *prettifier) prettify() error {
	for _, rule := range rules {
		err := rule(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *prettifier) code() string {
	p.buf.Reset()
	for e := p.tokenList.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
		p.buf.WriteString(e.Value.(*token.Token).Value())
	}
	return p.buf.String()
}

func getIndent(str string, n int) (indent string) {
	for n > 0 {
		indent = indent + str
		n--
	}
	return
}

func Prettify(tokens chan token.Token, w io.StringWriter, option *PrettyOption) error {
	if option == nil {
		option = DefaultPrettyOption
	}
	//var tokenArray []token.Token
	p := newPrettier(nil, option)
Loop:
	for {
		select {
		case t := <-tokens:
			//tokenArray = append(tokenArray, t)
			p.input(&t)
			if t.Kind == token.EOF {
				break Loop
			}
		}
	}
	err := p.prettify()
	if err != nil {
		return err
	}
	_, err = w.WriteString(p.code())
	if err != nil {
		return err
	}
	return nil
}
