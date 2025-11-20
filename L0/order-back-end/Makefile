include .env

LOCAL_BIN:=$(CURDIR)/bin

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

lint:
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.pipeline.yaml

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/golang-migrate/migrate/v4
	GOBIN=$(LOCAL_BIN) go install golang.org/x/tools/cmd/goimports@latest
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@v1.6.0


gen:
	$(LOCAL_BIN)/mockgen -source=internal/cache/cache.go \
	-destination=internal/cache/mocks/mock_cache.go
	$(LOCAL_BIN)/mockgen -source=internal/repository/repository.go \
	-destination=internal/repository/mocks/mock_repository.go

.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out