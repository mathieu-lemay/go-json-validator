MAKEFLAGS    += --always-make --warn-undefined-variables
SHELL        := /usr/bin/env bash
.SHELLFLAGS  := -e -o pipefail -c
.NOTPARALLEL :

PROJECT_ROOT ?= $(shell git rev-parse --show-toplevel)
# PROJECT_NAME ?= {{ cookiecutter.project_slug }}

format:
	pre-commit run --all-files golines

lint:
	pre-commit run --all-files

test:
	go test ./...

bench:
	go test -bench=. -run='^$$' ./...

mod:
	go mod tidy

update:
	go get -u -t ./...
	go mod tidy

clean:
	go clean
