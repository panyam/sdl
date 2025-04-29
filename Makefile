
watch:
	while true; do clear	; make run ; fswatch  -o ../ | echo "Files changed, re-testing..."; sleep 1 ; done

run:
	go test

test:
	go test ./...

bench:
	cd core && go test -bench=Benchmark -benchmem

testall: test bench
