include .env

# Go related variables.
GOBASE		:= $(shell pwd)
GOBIN		:= $(GOBASE)/build
PUBDIR     	:= $(GOBASE)/dist
TESTOUT		:= $(GOBIN)/coverage.out
TIMESTAMP	:= "$(shell date --rfc-3339='seconds')"
PKGS		:= $(shell go list ./... | grep -v /vendor)

ifeq (,$(wildcard .git))
REVER := $(shell svnversion -cn | sed -e "s/.*://" -e "s/\([0-9]*\).*/\1/" | grep "[0-9]")
else
REVER := $(shell git describe --always | sed -e "s/^v//")
endif

PROJECTBASE	:= $(GOBASE)/$(PROJPATH)
PROJECTNAME	?= $(shell basename "$(PROJECTBASE)")
OUTFILE		:= "$(PROJECTNAME)_$(shell go env GOARCH)_v$(PROJECTVER)-$(shell date +%Y%m%d%H%M%S)-$(REVER)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## build: compiling from source
.PHONY: build
build: 
	@echo ┌ compiling from source
	mkdir -p $(GOBIN)
	$(GO_ENV) go build $(GO_EXTRA_BUILD_ARGS) \
		-ldflags " \
			-X 'main.project="$(PROJECTNAME)"' \
			-X 'main.timestamp="$(TIMESTAMP)"' \
			-X 'main.version=$(PROJECTVER)' \
			-X 'main.revision=$(REVER)' \
		" \
		-o $(GOBIN)/$(PROJECTNAME) \
		$(PROJECTBASE)/main.go
	@echo └ done

## clean: cleanup workspace
.PHONY: clean
clean:
	@echo ┌ cleanup workspace
	@rm -rf $(GOBIN)
	@rm -rf $(PUBDIR)
	@rm -f $(TESTOUT)
	@echo └ done

## quick: run a quick case
.PHONY: quick
quick: build
	@echo "==> Run quick insert"
	- $(GOBIN)/$(PROJECTNAME) insert -f -r 15s

## test: running tests
.PHONY: test
test: 
	@echo ┌ running tests
	@rm -f $(TESTOUT)
	@go test -p 1 -v \
		-cover $(PKGS) \
		-coverprofile $(TESTOUT)
	@echo └ done

## lint: running code inspection
LINT_LOG := $(GOBASE)/build/lint.log
.PHONY: lint
lint:
	@echo ┌ running code inspection
	@echo │ "==>" run lint ...
	@echo lint result: > $(LINT_LOG)
	@for pkg in $(PKGS) ; do \
		golint $$pkg >> $(LINT_LOG); \
	done
	@echo │ "==>" run vet ...
	@echo vet result: >> $(LINT_LOG)
	- @go vet >> $(LINT_LOG) 2>&1 $(PKGS)
	@echo │ "===" found $$(grep -c ".go" $(LINT_LOG)) problems
	@echo │ "===" check $(LINT_LOG) for detail
	@echo └ inspection done

# shortcuts for development

.PHONY: requirements
requirements:
	@echo ┌ setup development requirements
	go install golang.org/x/lint/golint
	@echo └ done

## dist: build then package for release
.PHONY: dist
dist: build
	@echo ┌ compressing for release
	mkdir -p $(PUBDIR)
	tar --transform 's|.sample.|.|' \
		-czf $(PUBDIR)/$(OUTFILE).tar.gz \
		-C $(GOBIN) $(PROJECTNAME) \
		-C $(GOBASE) $(FILES)
	@echo └ done	
