
sdl:
	go build -o ./dist/sdl ./cmd/sdl/main.go

run:
	go test

test:
	go test ./...

bench:
	cd core && go test -bench=Benchmark -benchmem

watch:
	while true; do clear	; make run ; fswatch  -o ../ | echo "Files changed, re-testing..."; sleep 1 ; done

testall: test bench

prompt4sdl:
	source ~/personal/.shhelpers && files_for_llm `find . | grep -v "\.sh" | grep -v attic | grep -v mkprompt | grep -v parser | grep -v vscode | grep -v dsl | grep -v _test.go `

prompt4decl:
	source ~/personal/.shhelpers && files_for_llm `find . | grep -v "\..parser" | grep -v "\.sh" | grep -v attic | grep -v mkprompt | grep -v dsl  | grep -v vscode  | grep -v _test.go `
