.PHONY: lint
lint:
	-golint ./...
	go vet -v ./... || return 0

.PHONY: run
run:
	go run ./cmd/ports/... -c cmd/config.yml

.PHONY: build
build:
	go build -o build_output/ports ./cmd/ports

.PHONY: ports
ports:
	GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o build_output/ports ./cmd/ports