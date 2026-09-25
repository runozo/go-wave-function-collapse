os != go env GOOS
arch != go env GOARCH
distdirname := ./dist
distfilename := go-wave-function-collapse-${os}-${arch}
distfullpath := ${distdirname}/${distfilename}

.PHONY: wasm benchmark benchmark-compare test

wasm:
	GOOS=js GOARCH=wasm go build -ldflags="-s -w -v" -o ./docs/wfc.wasm github.com/runozo/go-wave-function-collapse

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
