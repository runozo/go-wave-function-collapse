os != go env GOOS
arch != go env GOARCH
distdirname := ./dist
distfilename := go-wave-function-collapse-${os}-${arch}
distfullpath := ${distdirname}/${distfilename}

.PHONY: wasm benchmark benchmark-compare test

wasm:
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ./docs/wfc.wasm github.com/runozo/go-wave-function-collapse
	@SRC="$$(go env GOROOT)/lib/wasm/wasm_exec.js"; \
	 [ -f "$$SRC" ] || SRC="$$(go env GOROOT)/misc/wasm/wasm_exec.js"; \
	 cp "$$SRC" ./docs/wasm_exec.js

build:
	go build -o ${distfullpath}

prod:
	go build -ldflags="-s -w -v" -o ${distfullpath}
	upx -9 ${distfullpath} --force-overwrite -o ${distfullpath}-packed

clean:
	rm -rf ${distdirname}

benchmark:
	BENCH_PATTERN=. ./scripts/bench.sh

# Compare the current tree against a git revision:
#   make benchmark-compare REV=a92eb05
benchmark-compare:
	@test -n "$(REV)" || { echo "usage: make benchmark-compare REV=<git-rev>"; exit 1; }
	./scripts/bench.sh $(REV)

test:
	go test ./...

headlessrun:
	go run main.go -iterations=2

run:
	go run main.go
