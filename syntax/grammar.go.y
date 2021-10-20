/*
  This parser definition is based on min-caml/parser.mly
  Copyright (c) 2005-2008, Eijiro Sumii, Moe Masuko, and Kenichi Asai
*/

%{
package syntax

import (
	"fmt"
	"strconv"
	"goscilla/ast"
	"goscilla/token"
)
%}

%union{
	node ast.Expr
	nodes []ast.Expr
	token *token.Token
	funcdef *ast.FuncDef
	decls []*ast.Symbol
	decl *ast.Symbol
	params []ast.Param
	program *ast.AST
}

%token<token> ILLEGAL

// whitespaces
%token<token> NEWLINE
%token<token> COMMENT
%token<token> WHITESPACE

// Literals
%token<token> STRING_LIT
%token<token> NUM_LIT
%token<token> HEX_LIT

// Prime Types
%token<token> INT_TYPE
%token<token> STRING_TYPE
%token<token> BYSTR_TYPE
%token<token> BNUM_TYPE
%token<token> MESSAGE_TYPE
%token<token> EVENT_TYPE

// Keywords
%token<token> FORALL
%token<token> BUILTIN
%token<token> LIBRARY
%token<token> IMPORT
%token<token> LET
%token<token> IN
%token<token> MATCH
%token<token> WITH
%token<token> END
%token<token> FUN
%token<token> TFUN
%token<token> CONTRACT
%token<token> TRANSITION
%token<token> SEND
%token<token> EVENT
%token<token> FIELD
%token<token> ACCEPT
%token<token> EXISTS
%token<token> DELETE
%token<token> EMP
%token<token> MAP
%token<token> SCILLA_VERSION
%token<token> TYPE
%token<token> OF
%token<token> TRY
%token<token> CATCH
%token<token> AS
%token<token> PROCEDURE
%token<token> THROW

// Separators
%token<token> SEMICOLON
%token<token> COLON
%token<token> PERIOD
%token<token> BAR
%token<token> LSQB
%token<token> RSQB
%token<token> LPAREN
%token<token> RPAREN
%token<token> LBRACE
%token<token> RBRACE
%token<token> COMMA
%token<token> ARROW
%token<token> TARROW
%token<token> EQ
%token<token> AND
%token<token> FETCH
%token<token> ASSIGN
%token<token> AT
%token<token> UNDERSCORE

// Identifiers
%token<token> ID   // simple name          [a-z][A-Za-z0-9_]*
%token<token> CID  // qualified name       [A-Z][A-Za-z0-9_]*
%token<token> TID  // type parameter name '[A-Z][A-Za-z0-9_]*
%token<token> SPID // special const name  _[A-Za-z0-9_]*

// from https://github.com/Zilliqa/scilla/blob/master/src/base/Datatypes.ml
// Builtin ADT
%token<token> BOOL
%token<token> TRUE
%token<token> FALSE

%token<token> NAT
%token<token> ZERO
%token<token> SUCC

%token<token> OPTION
%token<token> SOME
%token<token> NONE

%token<token> LIST
%token<token> CONS
%token<token> NIL

%token<token> PAIR

// Other tokens
%token<token> EOF

// Associativity
%right TARROW

%nonassoc IN
%right SEMICOLON
%nonassoc WITH
%right FETCH
%nonassoc BAR
%left COMMA
%left DOT
%nonassoc IDENT

%type<node> exp
%type<node> simple_exp
%type<node> seq_exp
%type<nodes> elems
%type<nodes> args
%type<params> params
%type<decls> pat
%type<funcdef> fundef
%type<token> match_arm_start
%type<decl> match_ident
%type<nodes> semi_elems
%type<node> type_annotation
%type<node> simple_type_annotation
%type<node> type
%type<node> simple_type
%type<node> simple_type_or_tuple
%type<nodes> arrow_types
%type<nodes> simple_type_star_list
%type<nodes> type_comma_list
%type<program> toplevels
%type<> opt_semi
%type<> program

%start program

%%
