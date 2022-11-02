# Go executable to use, i.e. `make GO=/usr/bin/go1.18`
# Defaults to first found in PATH
GO?=go

bin/bindl:
	export OUTDIR=bin/ && curl --location https://bindl.dev/bootstrap.sh | bash

include Makefile.*

.PHONY: bin/crewcli-dev
bin/crewcli-dev: bin/goreleaser
	([ -f bin/crewcli ] && rm bin/crewcli) || exit 0
	goreleaser build \
		--snapshot \
		--rm-dist \
		--single-target \
		--output bin/crewcli

.PHONY: release
release: bin/goreleaser bin/cosign


.PHONY: test/all
test:
	${GO} test -race -v ./...
