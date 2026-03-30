
all: clean test

parser.go: ./grammar.y ./lexer.go imports.go utils.go chainexpr.go
	goyacc -o parser.go -p "SDL" grammar.y

clean: 
	rm -f parser.go

test: parser.go
	go test

# build: goyacc -o parser.go -p "SDL" grammar.y
