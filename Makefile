os != go env GOOS
arch != go env GOARCH
distdirname := ./dist
distfilename := go-wave-function-collapse-${os}-${arch}
distfullpath := ${distdirname}/${distfilename}

.PHONY: wasm

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
	go test ./... -bench=.

test:
	go test ./...

headlessrun:
	go run main.go -iterations=2

run:
	go run main.go
