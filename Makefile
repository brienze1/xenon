SHELL := bash

TEST_THRESHOLD_COVERAGE := 65.0
MUTATION_THRESHOLD_COVERAGE := 72.0
MUTATION_THRESHOLD_EFFICACY := 100

WAILS_VERSION := v3.0.0-beta.28
GOIMPORTS_VERSION := v0.51.0
BETTERALIGN_VERSION := v0.15.2
DEADCODE_VERSION := v0.51.0
GOVULNCHECK_VERSION := v1.8.0
GREMLINS_VERSION := v0.6.0

TAG ?= @all

SANDBOX_ASSETS_DIR := internal/integration/webservice/sandbox/assets
PARITY_EXPECTED_DIR := test/integration/resources/scripts/corpus/expected
SNIPPET_EXPECTED_DIR := test/integration/resources/snippets/expected

GO_FILES = $(shell find . -name '*.go' -not -path './frontend/*' -not -path './build/*' -not -path './tools/*' -not -path './.git/*')
INTERNAL_PACKAGES = $(shell go list ./internal/... 2>/dev/null)
NO_INTERNAL_NOTE := "No packages under internal/ yet - skipping (the gate enforces once backend code exists)."
ifeq ($(OS),Windows_NT)
MUTATION_SKIP := true
MUTATION_SKIP_NOTE := "Mutation gates skipped on Windows: gremlins $(GREMLINS_VERSION) keys coverage by backslash paths and reports every mutant NOT COVERED there. macOS and Linux enforce them on the same code."
endif
MUTATION_SKIP_NOTE ?= "Mutation gates skipped (MUTATION_SKIP is set) - run make mutation-test locally to enforce them."

.PHONY: setup dev bindings build package ensure-dist sort-imports fix-struct-alignment vet \
	check-dead-code dependency-check tests bdd validate-coverage mutation-test \
	validate-mutation-coverage validate-test-efficacy frontend-check validate \
	sandbox-assets parity-oracle parity-diff parity script-reference snippet-golden newman

setup:
	@echo "Installing pinned tools..."
	go install github.com/wailsapp/wails/v3/cmd/wails3@$(WAILS_VERSION)
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install github.com/dkorunic/betteralign/cmd/betteralign@$(BETTERALIGN_VERSION)
	go install golang.org/x/tools/cmd/deadcode@$(DEADCODE_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	GOTOOLCHAIN=auto go install github.com/go-gremlins/gremlins/cmd/gremlins@$(GREMLINS_VERSION)
	@echo "Installing frontend dependencies..."
	cd frontend && npm ci
	@echo "Setup completed!"

dev:
	wails3 task dev

bindings:
	wails3 generate bindings -clean=true -ts -i ./cmd/xenon

build:
	wails3 build

package:
	wails3 package

ensure-dist:
	@if [ -z "$$(ls -A frontend/dist 2>/dev/null)" ]; then \
		mkdir -p frontend/dist && touch frontend/dist/.gitkeep; \
		echo "Created a placeholder frontend/dist for go:embed."; \
	fi

validate:
	@echo "Sort imports..."
	$(MAKE) sort-imports
	@echo "Fix struct alignment..."
	$(MAKE) fix-struct-alignment
	@echo "Running go vet..."
	$(MAKE) vet
	@echo "Running dead code analysis..."
	$(MAKE) check-dead-code
	@echo "Running dependency check..."
	$(MAKE) dependency-check
	@echo "Running unit tests..."
	$(MAKE) tests
	@echo "Validating test coverage..."
	$(MAKE) validate-coverage
	@echo "Running mutation tests..."
	$(MAKE) mutation-test
	@echo "Validating mutation coverage..."
	$(MAKE) validate-mutation-coverage
	@echo "Validating test efficacy..."
	$(MAKE) validate-test-efficacy
	@echo "Running frontend checks..."
	$(MAKE) frontend-check
	@echo "Validation completed!"

vet: ensure-dist
	@echo "Running go vet..."
	@go vet ./...
	@echo "No issues found by go vet."

sort-imports:
	@echo "Sorting imports..."
	@goimports -local github.com/brienze1/xenon -w $(GO_FILES)
	@echo "Imports sorted!"

fix-struct-alignment: ensure-dist
	@echo "Fixing struct alignment issues..."
	@while true; do \
		betteralign -apply ./...; \
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

check-dead-code: ensure-dist
	@echo "Running dead code analysis..."
	@dead_code_output=$$(deadcode -test ./... 2>&1); \
	if [ -n "$$dead_code_output" ]; then \
		echo "$$dead_code_output"; \
		echo ""; \
		echo "Dead code found:"; \
		exit 1; \
	else \
		echo "No dead code found."; \
	fi

dependency-check: ensure-dist
	@echo "Running govulncheck..."
	@govulncheck ./...
	@echo "Dependency check completed!"

tests: ensure-dist
	@echo "Running tests..."
	@if [ -n "$(INTERNAL_PACKAGES)" ]; then \
		go test ./... -coverpkg=./internal/... -coverprofile=coverage.out -p=100 && \
		go tool cover -func coverage.out; \
	else \
		echo $(NO_INTERNAL_NOTE) "(coverage profile)"; \
		rm -f coverage.out; \
		go test ./... -p=100; \
	fi
	@echo "Tests completed!"

bdd: ensure-dist
	@if [ -d test/integration/steps ]; then \
		go test ./test/integration/steps/... -v -count=1 -coverpkg=./internal/... -coverprofile=coverage.out --scenarios=$(TAG); \
	else \
		echo "No BDD suite in test/integration/steps yet - 0 scenarios."; \
	fi

validate-coverage:
	@echo "Validating coverage..."
	@if [ ! -f coverage.out ] || [ -z "$(INTERNAL_PACKAGES)" ]; then \
		echo $(NO_INTERNAL_NOTE); \
		exit 0; \
	fi; \
	statements=$$(awk 'NR > 1 { total += $$2 } END { print total + 0 }' coverage.out); \
	if [ "$$statements" -eq 0 ]; then \
		echo "No statements under internal/ yet - skipping coverage threshold."; \
		exit 0; \
	fi; \
	coverage=$$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $$NF}' | sed 's/%//'); \
	threshold=$(TEST_THRESHOLD_COVERAGE); \
	if awk -v c="$$coverage" -v t="$$threshold" 'BEGIN { exit !(c < t) }'; then \
		echo "Test coverage ($$coverage%) is below the required threshold ($$threshold%)."; \
		exit 1; \
	else \
		echo "Test coverage meets the required threshold ($$coverage% >= $$threshold%)."; \
	fi
	@echo "Coverage validation completed!"

mutation-test: ensure-dist
	@if [ -n "$(MUTATION_SKIP)" ]; then \
		echo $(MUTATION_SKIP_NOTE); \
		rm -f mutation_results.log; \
		exit 0; \
	fi; \
	if [ -z "$(INTERNAL_PACKAGES)" ]; then \
		echo $(NO_INTERNAL_NOTE); \
		rm -f mutation_results.log; \
		exit 0; \
	fi; \
	echo "Running mutation tests..."; \
	set -o pipefail; \
	XENON_SNIPPET_PROGRAMS=off gremlins unleash \
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
		--exclude-files "^build/" \
		--exclude-files "^frontend/" \
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
	@coverage=$$(grep -E 'Mutator coverage: [0-9]+(\.[0-9]+)?%' mutation_results.log 2>/dev/null | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ -z "$$coverage" ]; then \
		if [ -n "$(MUTATION_SKIP)" ]; then echo $(MUTATION_SKIP_NOTE); exit 0; fi; \
		if [ -z "$(INTERNAL_PACKAGES)" ]; then echo $(NO_INTERNAL_NOTE); exit 0; fi; \
		echo "No mutation coverage found in mutation_results.log - run make mutation-test first."; \
		exit 1; \
	fi; \
	threshold=$(MUTATION_THRESHOLD_COVERAGE); \
	if awk -v c="$$coverage" -v t="$$threshold" 'BEGIN { exit !(c < t) }'; then \
		echo "Mutation test coverage ($$coverage%) is below the required threshold ($$threshold%)."; \
		exit 1; \
	else \
		echo "Mutation test coverage meets the required threshold ($$coverage%)/($$threshold%)."; \
	fi
	@echo "Mutation coverage validation completed!"

validate-test-efficacy:
	@echo "Validating test efficacy..."
	@efficacy=$$(grep -E 'Test efficacy: [0-9]+(\.[0-9]+)?%' mutation_results.log 2>/dev/null | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ -z "$$efficacy" ]; then \
		if [ -n "$(MUTATION_SKIP)" ]; then echo $(MUTATION_SKIP_NOTE); exit 0; fi; \
		if [ -z "$(INTERNAL_PACKAGES)" ]; then echo $(NO_INTERNAL_NOTE); exit 0; fi; \
		echo "No test efficacy found in mutation_results.log - run make mutation-test first."; \
		exit 1; \
	fi; \
	threshold=$(MUTATION_THRESHOLD_EFFICACY); \
	if awk -v c="$$efficacy" -v t="$$threshold" 'BEGIN { exit !(c < t) }'; then \
		echo "Mutation test efficacy ($$efficacy%) is below the required threshold ($$threshold%)."; \
		exit 1; \
	else \
		echo "Mutation test efficacy meets the required threshold ($$efficacy%)/($$threshold%)."; \
	fi
	@echo "Test efficacy validation completed!"

frontend-check:
	cd frontend && npm run lint && npm run typecheck && npm test

sandbox-assets:
	@echo "Rebuilding the sandbox assets (postman-sandbox 6.7.4 libraries)..."
	cd tools/sandbox-assets && npm ci --no-audit --no-fund && node build.mjs && node build.mjs --check
	@if [ -n "$$(git status --porcelain -- $(SANDBOX_ASSETS_DIR))" ]; then \
		git status --short -- $(SANDBOX_ASSETS_DIR); \
		echo "Sandbox assets drifted from the committed build - review and commit $(SANDBOX_ASSETS_DIR)."; \
		exit 1; \
	fi
	@echo "Sandbox assets match MANIFEST.json and the committed build."

parity-oracle:
	@echo "Running the parity corpus through postman-sandbox 6.7.4..."
	cd tools/parity && npm ci --no-audit --no-fund && node oracle.mjs
	@if [ -n "$$(git status --porcelain -- $(PARITY_EXPECTED_DIR))" ]; then \
		git status --short -- $(PARITY_EXPECTED_DIR); \
		echo "Oracle outcomes changed - review and commit $(PARITY_EXPECTED_DIR)."; \
		exit 1; \
	fi
	@echo "Oracle outcomes match the committed corpus expectations."

parity-diff:
	TZ=UTC go test -tags parity -count=1 -v -run TestSandboxMatchesPostmanOracle ./test/integration/parity/... | grep -E -- "--- FAIL|differs|parity:|^(ok|FAIL)"; exit $${PIPESTATUS[0]}

parity: parity-oracle parity-diff

snippet-golden:
	@echo "Generating the snippet golden corpus with postman-code-generators 2.1.1..."
	cd tools/snippets && npm ci --no-audit --no-fund && node golden.mjs
	@if [ -n "$$(git status --porcelain -- $(SNIPPET_EXPECTED_DIR))" ]; then \
		git status --short -- $(SNIPPET_EXPECTED_DIR); \
		echo "Snippet golden outputs changed - review and commit $(SNIPPET_EXPECTED_DIR)."; \
		exit 1; \
	fi
	@echo "Snippet golden outputs match the committed corpus."

script-reference:
	@echo "Generating the script editor reference from reference/members.tsv and the sandbox shims..."
	go test ./internal/integration/webservice/sandbox/ -run TestScriptReferenceIsGenerated -count=1 -update-reference
	@echo "Reference written to internal/integration/webservice/sandbox/reference/reference.json."

newman: ensure-dist
	@echo "Installing the pinned Newman (tools/newman)..."
	cd tools/newman && npm ci --no-audit --no-fund
	@echo "Exporting the Xenon-created corpus to Postman and running it with Newman against the mock server..."
	XENON_NEWMAN=on go test ./test/integration/steps/... -v -count=1 --scenarios=@newman
