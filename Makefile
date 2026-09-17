.PHONY: fmt vet unit race coverage gosec govulncheck acceptance acceptance-mutation crap verify install

export PATH := $(CURDIR)/.tools/bin:$(PATH)
PKGS := $(shell go list ./... | grep -v '/acceptance/generated')

fmt:
	@test -z "$$(gofmt -l . | grep -v /acceptance/generated/)"

vet:
	go vet $(PKGS)

unit:
	go test $(PKGS)

race:
	go test -race $(PKGS)

coverage:
	go test -coverprofile=coverage.out $(PKGS)

gosec:
	gosec -exclude-dir=.tools -exclude-dir=acceptance/generated ./...

govulncheck:
	govulncheck ./...

acceptance:
	sh scripts/acceptance.sh

acceptance-mutation:
	sh scripts/acceptance-mutation.sh

crap:
	sh scripts/crap.sh

verify: fmt vet unit race coverage gosec govulncheck acceptance crap

install:
	go install ./cmd/m365
