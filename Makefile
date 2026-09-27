.PHONY: gomod fmt vet unit race coverage gosec govulncheck acceptance acceptance-mutation crap verify install

TOOLS_BIN := $(shell pwd)/.tools/bin
# Run every target with the toolchain CI uses (go.mod's toolchain line),
# whatever Go is installed. Override with GOTOOLCHAIN=local to test a newer Go.
# An unset or empty GOTOOLCHAIN gets the pin, as in install-tools.sh.
ifeq ($(strip $(GOTOOLCHAIN)),)
GOTOOLCHAIN := $(shell sed -n 's/^toolchain //p' go.mod)
endif
export GOTOOLCHAIN
# $(shell) does not see exported variables, so pass the toolchain explicitly.
PKGS := $(shell GOTOOLCHAIN=$(GOTOOLCHAIN) go list ./... | grep -v '/acceptance/generated')

gomod:
	sh scripts/gomod.sh

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
	PATH="$(TOOLS_BIN):$$PATH" gosec -exclude-dir=.tools -exclude-dir=acceptance/generated ./...

govulncheck:
	PATH="$(TOOLS_BIN):$$PATH" govulncheck ./...

acceptance:
	sh scripts/acceptance.sh

acceptance-mutation:
	sh scripts/acceptance-mutation.sh

crap:
	sh scripts/crap.sh

verify: gomod fmt vet unit race coverage gosec govulncheck acceptance crap

install:
	go install ./cmd/m365
