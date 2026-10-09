---
name: Makefile Validation Flow
description: This skill should be used when the user asks about "make validate", "validation pipeline", "import sorting", "struct alignment", "dead code", "mutation testing", "coverage threshold", "make targets", or when verifying changes pass all quality gates.
version: 1.0.0
---

# Makefile — Validation Pipeline and Build Targets

The project root `Makefile` defines the full validation pipeline and build targets. The orchestrator or tester agent should run `make validate` (or individual steps) to verify changes.

---

## Master Validation Command

```bash
make validate
```

This runs ALL quality checks in order. If any step fails, the pipeline stops.

---

## Validation Pipeline Steps (in order)

### 1. `make sort-imports`

Uses `goimports` to auto-sort and organize all Go imports. Fixes IDE warnings about import order.

**Import ordering convention:** stdlib first, then external packages, then internal packages. `goimports` handles this automatically.

```bash
make sort-imports
```

### 2. `make fix-struct-alignment`

Uses `betteralign` to optimize struct field alignment for memory efficiency. Runs in a loop until all fixes are applied.

```bash
make fix-struct-alignment
```

### 3. `make vet`

Runs `go vet ./...` for static analysis.

```bash
make vet
```

### 4. `make check-dead-code`

Uses `deadcode` to find unused code. Fails if any dead code is found.

```bash
make check-dead-code
```

### 5. `make dependency-check`

Uses `nancy` for vulnerability scanning of dependencies. Requires `NANCY_USERNAME` and `NANCY_TOKEN` environment variables.

```bash
make dependency-check
```

### 6. `make tests`

Runs all tests with coverage profiling against `./internal/...`.

```bash
make tests
```

### 7. `make validate-coverage`

Checks that test coverage meets the threshold (currently **65.0%**).

```bash
make validate-coverage
```

### 8. `make mutation-test`

Runs mutation testing with `gremlins`. Excludes `cmd/`, `pkg/`, `test/`, `config/`, `build/`, `.claude/` directories. Covers `./internal/...` only.

```bash
make mutation-test
```

### 9. `make validate-mutation-coverage`

Checks that mutation coverage meets the threshold (currently **72.0%**).

```bash
make validate-mutation-coverage
```

### 10. `make validate-test-efficacy`

Checks that test efficacy meets the threshold (currently **100%**).

```bash
make validate-test-efficacy
```

---

## Other Targets

| Target | Description |
|---|---|
| `make run` | Runs the app locally via `cmd/local/main.go` |
| `make build-binary` | Cross-compiles for Linux ARM64 (Lambda deployment) |
| `make tests-with-tag` | Runs integration tests with tag filtering (`--scenarios=@all`) and generates HTML coverage |
| `make go-list` | Lists all modules in JSON format |
| `make fix-test-env-encoding` | Fixes encoding of test `.env` file (ISO-8859-1 to UTF-8) |
| `make fix-env-encoding` | Fixes encoding of root `.env` file |

---

## When to Use

- **After making code changes:** Run `make sort-imports` to fix import ordering before running tests.
- **Before final verification:** Run `make validate` for the full static validation pipeline.
- **During development:** Run individual targets (`make vet`, `make tests`, `make check-dead-code`) for quick feedback.
- **After adding/removing code:** Run `make check-dead-code` to catch unused functions.
- **After changing struct fields:** Run `make fix-struct-alignment` to optimize memory layout.

---

## Thresholds

| Metric | Threshold | Variable |
|---|---|---|
| Test coverage | 65.0% | `TEST_THRESHOLD_COVERAGE` |
| Mutation coverage | 72.0% | `MUTATION_THRESHOLD_COVERAGE` |
| Test efficacy | 100% | `MUTATION_THRESHOLD_EFFICACY` |

---

## Makefile Template

```makefile
TEST_THRESHOLD_COVERAGE := 65.0
MUTATION_THRESHOLD_COVERAGE := 72.0
MUTATION_THRESHOLD_EFFICACY=100

.PHONY: run build-binary validate deadcode vet

run:
	go run ./cmd/local/main.go

build-binary:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o ${app_name} -ldflags="-s -w" ./cmd/app/main.go

go-list:
	go list -modfile=go.mod -m -json -mod=mod all

fix-test-env-encoding:
	iconv -f ISO-8859-1 -t UTF-8 ./test/integration/resources/.env.test > ./test/integration/resources/.env.tmp && mv ./test/integration/resources/.env.tmp ./test/integration/resources/.env.test

fix-env-encoding:
	iconv -f ISO-8859-1 -t UTF-8 .env > .env.tmp && mv .env.tmp .env

validate:
	@echo "Sort imports..."
	make sort-imports
	@echo "Fix struct alignment..."
	make fix-struct-alignment
	@echo "Running go vet..."
	make vet
	@echo "Running dead code analysis..."
	make check-dead-code
	@echo "Running dependency check..."
	make dependency-check
	@echo "Running unit tests..."
	make tests
	@echo "Validating test coverage..."
	make validate-coverage
	@echo "Running mutation tests..."
	make mutation-test
	@echo "Validating mutation coverage..."
	make validate-mutation-coverage
	@echo "Validating test efficacy..."
	make validate-test-efficacy

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "No issues found by go vet."

sort-imports:
	@echo "Sorting imports..."
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .
	@echo "Imports sorted!"

fix-struct-alignment:
	@echo "Installing betteralign..."
	@go install github.com/dkorunic/betteralign/cmd/betteralign@latest
	@echo "Fixing struct alignment issues..."
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then GOBIN="$$(go env GOPATH)/bin"; fi; \
	while true; do \
		"$$GOBIN/betteralign" -apply ./...; \
		exit_code=$$?; \
		if [ $$exit_code -eq 0 ]; then \
			break; \
		elif [ $$exit_code -eq 3 ]; then \
			echo "Running betteralign again to apply remaining fixes..."; \
		else \
			echo "Error: betteralign failed with exit code $$exit_code"; \
			exit $$exit_code; \
		fi; \
	done
	@echo "Struct alignment fixed!"

check-dead-code:
	@echo "Installing deadcode..."
	@go install golang.org/x/tools/cmd/deadcode@latest
	@echo "Running dead code analysis..."
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then GOBIN="$$(go env GOPATH)/bin"; fi; \
	dead_code_output=$$("$$GOBIN/deadcode" -test ./... 2>&1); \
	if [ -n "$$dead_code_output" ]; then \
		echo "$$dead_code_output"; \
		echo ""; \
		echo "Dead code found:"; \
		exit 1; \
	else \
		echo "No dead code found."; \
	fi

dependency-check: autofix-dependency-check
	@echo "Installing Nancy..."
	go install github.com/sonatype-nexus-community/nancy@latest
	go mod tidy
	@echo "Nancy installed successfully!"
	@echo "Running dependency check..."
	@if [ -z "$$NANCY_USERNAME" ] || [ -z "$$NANCY_TOKEN" ]; then \
		echo "Error: NANCY_USERNAME and NANCY_TOKEN environment variables must be set"; \
		exit 1; \
	fi
	@bash -o pipefail -c 'go list -json -deps ./... | nancy sleuth --username "$$NANCY_USERNAME" --token "$$NANCY_TOKEN" --exclude-vulnerability-file .nancy-ignore | tee nancy-report.log'
	@echo "Dependency check completed!"

autofix-dependency-check:
	@echo "Running dependency check fixes..."
	findings_file="nancy-report.log"; \
	grep -Eo 'pkg:golang/[a-zA-Z0-9._/-]+@[v0-9]+\.[0-9]+\.[0-9]+' $$findings_file | while read -r line; do \
		module=$$(echo $$line | sed 's/pkg:golang\///' | awk -F'@' '{print $$1}'); \
		echo "Updating $$module to the latest version..."; \
		go get -u "$$module"; \
	done; \
	go mod tidy
	@echo "Dependency check fix completed!"

tests:
	@echo "Running tests..."
	go test ./... -coverpkg=./internal/... -coverprofile=coverage.out ./... -p=100
	go tool cover -func coverage.out
	@echo "Tests completed!"

tests-with-tag:
	@echo "Running tests..."
	go test ./test/integration/... -v -coverprofile=coverage.out -coverpkg=./internal/... -p 100 --scenarios=@all
	go tool cover -html=coverage.out -o coverage.html
	@echo "Tests completed!"

validate-coverage:
	@echo "Validating coverage..."
	@coverage=$$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $$NF}' | sed 's/%//'); \
	threshold=$(TEST_THRESHOLD_COVERAGE); \
	if (( $$(echo "$$coverage < $$threshold" | bc -l) )); then \
		echo "Test coverage ($$coverage%) is below the required threshold ($$threshold%)."; \
		exit 1; \
	else \
		echo "Test coverage meets the required threshold ($$coverage% >= $$threshold%)."; \
	fi
	@echo "Coverage validation completed!"

mutation-test:
	@echo "Installing Gremlins..."
	GOTOOLCHAIN=auto go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
	@echo "Gremlins installed successfully!"
	@echo "Running mutation tests..."
	gremlins unleash \
              --silent=false \
              --integration=true \
              --dry-run=false \
              --tags="" \
              --output="" \
              --workers=4 \
              --test-cpu=2 \
              --timeout-coefficient=0 \
              --threshold-efficacy=90 \
              --threshold-mcover=50 \
              --exclude-files "^cmd/" \
              --exclude-files "^pkg/" \
              --exclude-files "^test/" \
              --exclude-files "^config/" \
              --exclude-files "^build/" \
              --exclude-files "^.claude/" \
              --coverpkg "./internal/..." \
              --arithmetic-base=true \
              --conditionals-boundary=true \
              --conditionals-negation=true \
              --increment-decrement=true \
              --invert-assignments=true \
              --invert-bitwise=true \
              --invert-bwassign=true \
              --invert-negatives=true \
              --invert-logical=true \
              --invert-loopctrl=true \
              --remove-self-assignments=true | tee mutation_results.log
	@echo "Mutation tests completed!"

validate-mutation-coverage:
	@echo "Validating mutation coverage..."
	coverage=$$(grep -E 'Mutator coverage: [0-9]+(\.[0-9]+)?%' mutation_results.log | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	threshold=$(MUTATION_THRESHOLD_COVERAGE); \
	result=$$(echo "$$coverage >= $$threshold" | bc); \
	if [ "$$result" -eq 0 ]; then \
		echo "Mutation test coverage ($$coverage%) is below the required threshold ($$threshold%)."; \
		exit 1; \
	else \
		echo "Mutation test coverage meets the required threshold ($$coverage%)/($$threshold%)."; \
	fi
	@echo "Mutation coverage validation completed!"

validate-test-efficacy:
	@echo "Validating test efficacy..."
	efficacy=$$(grep -E 'Test efficacy: [0-9]+(\.[0-9]+)?%' mutation_results.log | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	threshold=$(MUTATION_THRESHOLD_EFFICACY); \
	result=$$(echo "$$efficacy >= $$threshold" | bc); \
	if [ "$$result" -eq 0 ]; then \
		echo "Mutation test coverage ($$efficacy%) is below the required threshold ($$threshold%)."; \
		exit 1; \
	else \
		echo "Mutation test efficacy meets the required threshold ($$efficacy%)/($$threshold%)."; \
	fi
	@echo "Mutation coverage validation completed!"
```

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
