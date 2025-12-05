
parser.go: ./grammar.y ./lexer.go imports.go utils.go chainexpr.go
	goyacc -o parser.go -p "SDL" grammar.y

all: test

test: parser
	go test

# build: goyacc -o parser.go -p "SDL" grammar.y
