package main

import (
	"flag"
	"fmt"
	"github.com/rhysd/locerr"
	"github.com/sirupsen/logrus"
	"goscilla/driver"
	"os"
)

var (
	help       = flag.Bool("help", false, "Show this help")
	showTokens = flag.Bool("tokens", false, "Show tokens for input")
	showAST    = flag.Bool("ast", false, "Show AST for input")
	check      = flag.Bool("check", false, "Check code (syntax, types, ...) and report errors if exist")
)

const usageHeader = `Usage: gocaml [flags] [file]

  Compiler for GoCaml.
  When file is given as argument, compiler will compile it. Otherwise, compiler
  attempt to read from STDIN as source code to compile.

Flags:`

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, usageHeader)
	flag.PrintDefaults()
}

func init() {
	l, _ := logrus.ParseLevel(os.Getenv("GOSCILLA_LOG_LEVEL"))
	logrus.SetLevel(l)
	logrus.SetOutput(os.Stderr)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		os.Exit(0)
	}

	var src *locerr.Source
	var err error

	if flag.NArg() == 0 {
		src, err = locerr.NewSourceFromStdin()
	} else {
		src, err = locerr.NewSourceFromFile(flag.Arg(0))
	}

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error on opening source: %s\n", err.Error())
		os.Exit(4)
	}

	d := driver.Driver{}

	switch {
	case *showTokens:
		d.PrintTokens(src)
	case *showAST:
	case *check:
	default:
		//d.PrintAST(src)
		d.Prettify(src)
	}
}
