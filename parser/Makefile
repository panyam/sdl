
all: build test

test:
	go test

build:
	goyacc -o parser.go -p "SDL" grammar.y
