.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/app/grpc/pb/shortener.proto

.PHONY: install-tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: build
build:
	go build -o bin/shortener cmd/shortener/main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	go vet ./...
	go run cmd/staticlint/main.go ./...
