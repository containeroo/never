# Detect platform for sed compatibility
SED := $(shell if [ "$(shell uname)" = "Darwin" ]; then echo gsed; else echo sed; fi)

# VERSION defines the project version, extracted from cmd/portpatrol/main.go without leading 'v'.
VERSION := $(shell awk -F'"' '/const version/{gsub(/^v/, "", $$2); print $$2}' cmd/portpatrol/main.go)

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.1.2

.PHONY: test cover clean update patch minor major tag

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: download
download: ## Download go packages
	go mod download

.PHONY: update-packages
update-packages: ## Update all Go packages to their latest versions
	go get -u ./...
	go mod tidy

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: ## Run all unit tests
	go test -coverprofile=coverage.out -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...

.PHONY: cover
cover: ## Display test coverage
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## Clean up generated files
	rm -f coverage.out coverage.html

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	$(GOLANGCI_LINT) run --fix

##@ Versioning

patch: ## Increment the patch version (x.y.Z -> x.y.(Z+1)).
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{print $$1"."$$2"."$$3+1}') && \
	$(SED) -i -E "s/(const version string = \"v)[^\"]+/\1$${NEW_VERSION}/" cmd/portpatrol/main.go

minor: ## Increment the minor version (x.Y.z -> x.(Y+1).0).
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{print $$1"."$$2+1".0"}') && \
	$(SED) -i -E "s/(const version string = \"v)[^\"]+/\1$${NEW_VERSION}/" cmd/portpatrol/main.go

major: ## Increment the major version (X.y.z -> (X+1).0.0).
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{print $$1+1".0.0"}') && \
	$(SED) -i -E "s/(const version string = \"v)[^\"]+/\1$${NEW_VERSION}/" cmd/portpatrol/main.go

tag: ## Tag the current commit with the current version if no tag exists and the repository is clean.
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Repository has uncommitted changes. Please commit or stash them before tagging."; \
		exit 1; \
	fi
	@if [ -z "$$(git tag --list v$(VERSION))" ]; then \
		echo "Tagging version v$(VERSION)"; \
		git tag "v$(VERSION)"; \
		git push origin "v$(VERSION)"; \
	else \
		echo "Tag v$(VERSION) already exists."; \
	fi


##@ Dependencies

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

