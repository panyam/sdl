
watch:
	while true; do clear	; make run ; fswatch  -o ../ | echo "Files changed, re-testing..."; sleep 1 ; done

run:
	go test

test:
	go test ./...

bench:
	cd core && go test -bench=Benchmark -benchmem

testall: test bench

prompt:
	source ~/personal/.shhelpers && files_for_llm `find . | grep -v "\.sh" | grep -v attic | grep -v mkprompt | grep sdl `
