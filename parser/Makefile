
all: build test

test:
	go test

build:
	goyacc -o parser.go -p "yy" grammar.y
