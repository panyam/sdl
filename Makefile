
binary: gen
	cd parser && make
	go build -o ${GOBIN}/sdl ./cmd/sdl/main.go

gen:
	buf generate

# Development workflow: build and test dashboard
dev-test: binary
	cd web && ./dev-test.sh

# Quick development validation
dev-quick: binary
	cd web && npm run dev-quick

dev-screenshot: binary
	cd web && npm run dev-screenshot

run:
	go test

test:
	go test ./...

bench:
	cd core && go test -bench=Benchmark -benchmem

watch:
	while true; do clear	; make run ; fswatch  -o ../ | echo "Files changed, re-testing..."; sleep 1 ; done

testall: test bench

sdlfiles:
	@find . | grep -v "\.git" | grep -v "\.sh" | grep -v "\..decl" | grep -v attic | grep -v prompt | grep -v vscode | grep -v dsl | grep -v _test.go | grep -v "\.bak" | grep -v debug | grep -v "\.output" | grep -v "\.svg" | grep -v "\.png" | grep -v parser.go | grep -v CLAUDE | grep -v node_modules | grep -v "\.\/sdl" | grep -v package | grep -v playwright-report | grep -v test-results | grep -v dist | grep -v sdl

sdlprompt:
	source ~/personal/.shhelpers && files_for_llm `make sdlfiles`

parserfiles:
	@find . |  grep -v "\.git" | grep -v "\.bak" | grep -e "decl" -e "parser" | grep -v ll | grep -v y.output | grep -v parser.go  | grep -v debug | grep -v "\.output"

parserprompt:
	source ~/personal/.shhelpers && files_for_llm `make parserfiles`

newchatprompt:
	cat prompts/STARTER.prompt && make sdlprompt
