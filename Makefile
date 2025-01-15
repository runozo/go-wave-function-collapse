os != go env GOOS
arch != go env GOARCH
distdirname := ./dist
distfilename := go-wave-function-collapse-${os}-${arch}
distfullpath := ${distdirname}/${distfilename}

.PHONY: prod

prod:
	go build -ldflags="-s -w -v" -o ${distfullpath}
	upx -9 ${distfullpath} --force-overwrite -o ${distfullpath}-packed

.PHONY: build

build:
	go build -o ${distfullpath} ./

.PHONY: run

run: build
	${distfullpath}