GO := go
srcs := $(shell find . -path ./vendor -prune -o -name '*.go' | grep -v 'vendor')
PKGS ?= $(shell go list ./...)
PKG_FILES ?= *.go

RACE=-race
GOTEST=go test -v $(RACE)
GOLINT=golint
GOVET=go vet
GOFMT=gofmt
FMT_LOG=fmt.log
LINT_LOG=lint.log

.PHONY: dependencies
dependencies:
	@echo "Installing golang dep if needed and looking for dependencies"
	go mod download
	go get -u github.com/axw/gocov/gocov
	go get -u golang.org/x/lint/golint

.PHONY: fmt
fmt:
	$(GOFMT) -e -s -l -w $(srcs)

.PHONY: lint
lint:
	@rm -rf $(LINT_LOG)
	@rm -rf $(FMT_LOG)
	@echo "gofmt the files..."
	$(GOFMT) -e -s -l -w $(srcs) > $(FMT_LOG)
	@[ ! -s "$(FMT_LOG)" ] || (echo "Go Fmt Failures, run 'make fmt'" | cat - $(FMT_LOG) && false)
	@echo "Installing test dependencies for vet..."
	@go test -i $(PKGS)
	@echo "Checking vet..."
	$(GOVET) $(PKGS)
	@echo "Checking lint..."
	@$(foreach dir,$(PKGS),golint $(dir) 2>&1 | tee -a $(LINT_LOG);)
	@echo "Checking for unresolved FIXMEs..."
	@git grep -i fixme | grep -v -e Makefile | tee -a $(LINT_LOG)
	@[ ! -s "$(LINT_LOG)" ] || (echo "Lint Failures" | cat - $(LINT_LOG) && false)


.PHONY: test
test:
	$(GOTEST) $(PKGS)
