PROG=ovh-ddns
OUT=bin/$(PROG)
MAIN=cmd/main.go
COVERAGE_FILE=test/coverage.out
COVERAGE_HTML_FILE=test/coverage.html
TEST_REPORT=test/test-report.json

.PHONY: all
all: build

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(OUT) $(MAIN)

.PHONY: image
image:
	@docker build --target runtime -t zoui/$(PROG):latest .

.PHONY: compose-up
compose-up:
	@mkdir -p test/data
	@docker compose -f test/compose.yaml up

.PHONY: compose-down
compose-down:
	@docker compose -f test/compose.yaml down

.PHONY: run
run:
	@go run -tags dev $(MAIN)

.PHONY: test
test:
	@mkdir -p test
	@go test -tags dev -cover -coverprofile=$(COVERAGE_FILE) ./...

.PHONY: test-cicd
test-cicd:
	@mkdir -p test
	@go test -tags dev -v -race -cover -coverprofile=$(COVERAGE_FILE) ./...

.PHONY: test-race
test-race:
	@mkdir -p test
	@go test -tags dev -race ./...

.PHONY: benchmark
benchmark:
	@go test -tags dev -bench=. -benchmem -run =^a ./...

.PHONY: coverage
coverage: test
	@go tool cover -html=$(COVERAGE_FILE) -o=$(COVERAGE_HTML_FILE)

.PHONY: clean
clean:
	@rm -rf bin

.PHONY: gitclean
gitclean:
	@git clean -xdf
