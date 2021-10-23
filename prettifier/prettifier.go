package prettifier

import (
	"bytes"
	"container/list"
	"goscilla/token"
	"io"
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

type prettifier struct {
	tokens    []token.Token
	tokenList list.List
	contexts  []token.Kind
	newLines  int
	lastLine  string
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

func (p *prettifier) input(t *token.Token) {
	p.tokenList.PushBack(t)
}

func (p *prettifier) reset() {
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
		//logrus.Traceln(e.Value)
		p.buf.WriteString(e.Value.(*token.Token).Value())
	}
	return p.buf.String()
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
