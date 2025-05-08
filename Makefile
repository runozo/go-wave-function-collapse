os != go env GOOS
arch != go env GOARCH
distdirname := ./dist
distfilename := go-wave-function-collapse-${os}-${arch}
distfullpath := ${distdirname}/${distfilename}

.PHONY: prod

prod:
	go build -ldflags="-s -w -v" -o ${distfullpath}
	upx -9 ${distfullpath} --force-overwrite -o ${distfullpath}-packed

.PHONY: clean

clean:
	rm -rf ${distdirname}

.PHONY: benchmark

benchmark:
	go test ./... -bench=.

.PHONY: test

test:
	go test ./...

.PHONY: run

run:
	go run main.go
