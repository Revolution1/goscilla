
all:

syntax/grammar.go: syntax/grammar.go.y
	#go get golang.org/x/tools/cmd/goyacc
	goyacc -o syntax/grammar.go syntax/grammar.go.y