package syntax

import (
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var (
	log                *logrus.Logger
	traceLexerPos      bool
	traceLexerToken    bool
	traceLexerTokenPos bool
	traceParser        bool
)

func init() {
	log = logrus.StandardLogger()
	traceLexerPos, _ = strconv.ParseBool(os.Getenv("TRACE_LEXER_POS"))
	traceLexerToken, _ = strconv.ParseBool(os.Getenv("TRACE_LEXER_TOKEN"))
	traceLexerTokenPos, _ = strconv.ParseBool(os.Getenv("TRACE_LEXER_TOKEN_POS"))
	traceParser, _ = strconv.ParseBool(os.Getenv("TRACE_PARSER"))
}
