# Copyright 2026 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Repository development targets and pinned static analysis tools.
.DEFAULT_GOAL := help
.PHONY: fix git-hooks help test test-race verify

help:
	@printf '  %-14s - %s\n' \
		'make fix' 'Apply formatting and safe lint fixes' \
		'make verify' 'Run static checks' \
		'make test' 'Run Go tests with coverage and benchmark smoke checks' \
		'make test-race' 'Run Go tests with the race detector' \
		'make git-hooks' 'Install versioned Git hooks'

fix:
	@uv run --locked ruff check --quiet --fix scripts
	@uv run --locked ruff format scripts >/dev/null
	@uv run --locked python scripts/update_agents_file_index.py
	@go tool golangci-lint fmt
	@go tool golangci-lint run --show-stats=false --fix ./...

verify:
	@uv run --locked ruff check --quiet scripts
	@uv run --locked ruff format --check --quiet scripts
	@go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
	@go tool golangci-lint fmt --diff
	@go tool golangci-lint run --show-stats=false ./...
	@go test -run '^$$' ./...
	@go tool addlicense -ignore '.venv/**' -ignore '.ruff_cache/**' -ignore 'third_party/**' -ignore 'vendor/**' -check .
	@uv run --locked python scripts/lint_binaries.py
	@uv run --locked python scripts/update_agents_file_index.py --check
	# Module metadata contains upstream dependency names rather than prose.
	@git ls-files -z -- . ':(exclude)go.mod' ':(exclude)go.sum' | xargs -0 go run github.com/client9/misspell/cmd/misspell@v0.3.4 -error
	@go mod tidy -diff

test:
	@go test -timeout=120s -covermode=count -coverprofile coverage.txt -bench=. -benchtime=1x ./...

test-race:
	@go test -timeout=120s -race -bench=. -benchtime=1x ./...

git-hooks:
	@./scripts/install-git-hooks.sh
