# Go executable to use, i.e. `make GO=/usr/bin/go1.18`
# Defaults to first found in PATH
GO?=go

bin/bindl:
	export OUTDIR=bin/ && curl --location https://bindl.dev/bootstrap.sh | bash

include Makefile.*

.PHONY: bin/crewcli-dev
bin/crewcli-dev: bin/goreleaser
	([ -f bin/crewcli ] && rm bin/crewcli) || exit 0
	GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) \
		goreleaser build \
			--snapshot \
			--rm-dist \
			--single-target \
			--output bin/crewcli

.PHONY: release
release: bin/goreleaser bin/cosign
	PATH=${PWD}/bin:${PATH} bin/goreleaser release --rm-dist

.PHONY: test/all
test:
	${GO} test -race -v ./...

.PHONY: lint
lint: bin/golangci-lint
	bin/golangci-lint run

.PHONY: lint/fix
lint/fix: bin/golangci-lint
	bin/golangci-lint run --fix

.PHONY: lint/gh-actions
lint/gh-actions: bin/golangci-lint
	bin/golangci-lint run --out-format github-actions
