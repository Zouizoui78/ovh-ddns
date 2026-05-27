PROG=ovh-ddns
OUT=bin/$(PROG)
MAIN=cmd/main.go
COVERAGE_FILE=test/coverage.out
TEST_REPORT=test/test-report.json

.PHONY: all
all: build

.PHONY: build
build:
	@go build -ldflags="-s -w" -o $(OUT) $(MAIN)

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
test: swagger
	@go test -tags dev -cover -coverprofile=$(COVERAGE_FILE) ./...

.PHONY: test-cicd
test-cicd: swagger
	@go test -tags dev -v -race -cover -coverprofile=$(COVERAGE_FILE) -json ./... > $(TEST_REPORT)

.PHONY: benchmark
benchmark:
	@go test -tags dev -bench=. -benchmem -run =^a ./...

.PHONY: coverage
coverage: test
	@go tool cover -html=$(COVERAGE_FILE)

.PHONY: clean
clean:
	@rm -rf bin

.PHONY: gitclean
gitclean:
	@git clean -xdf
