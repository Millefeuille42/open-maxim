.PHONY: build test clean docker-build docker-run linter lint format formatter lint-fix fmt

APP_NAME = maxim_server
MAIN_PATH = ./cmd/server

build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker run --rm -ti -p8080:8080 -p2002:2002 $(APP_NAME):latest

./bin/golangci-lint:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.14.0

linter: ./bin/golangci-lint
formatter: ./bin/golangci-lint

lint: linter
	./bin/golangci-lint run ./cmd/... ./internal/...
lint-fix: linter
	./bin/golangci-lint run ./cmd/... ./internal/... --fix
format: formatter
	./bin/golangci-lint fmt ./cmd/... ./internal/...
fmt: format