.PHONY: prod

prod:
	go build -ldflags="-s -w -v" -o ./go-wave-function-collapse .
	upx -9 ./go-wave-function-collapse

.PHONY: build

build:
	go build -o ./go-wave-function-collapse .

.PHONY: run

run: build
	./go-wave-function-collapse