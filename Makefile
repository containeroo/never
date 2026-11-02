.DEFAULT_GOAL := help

# renovate: datasource=github-releases depName=gi8lino/dev-tools
DEV_TOOLS_VERSION ?= v0.9.0

include bin/dev-tools.mk
include $(call dev-tools-module,tag)
include $(call dev-tools-module,help)

LOCALBIN ?= $(CURDIR)/bin

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Documentation Tools
KUMBUKA_CLI ?= $(LOCALBIN)/kumbuka-cli

## Tool Versions
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.13.2
# renovate: datasource=github-releases depName=kumbuka-me/cli
KUMBUKA_CLI_VERSION ?= v0.13.10

## Site Configuration
SITE_CONFIG ?= docs/kumbuka-site.toml
SITE_OUTPUT ?= docs/site
SITE_PORT ?= 8081
SITE_LOGO ?= docs/assets/logo.svg
SITE_FAVICON_SVG ?= docs/content/assets/favicon.svg
SITE_FAVICON_ICO ?= docs/content/favicon.ico
MAGICK ?= magick

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
test: fmt vet ## Run unit tests.
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

.PHONY: site-favicon
site-favicon: ## Generate the documentation SVG favicon and multi-size ICO (requires ImageMagick).
	@command -v "$(MAGICK)" >/dev/null 2>&1 || { echo "ImageMagick is required; install it and set MAGICK if needed." >&2; exit 1; }
	@test -f "$(SITE_LOGO)" || { echo "SVG logo not found: $(SITE_LOGO)" >&2; exit 1; }
	mkdir -p "$(dir $(SITE_FAVICON_SVG))" "$(dir $(SITE_FAVICON_ICO))"
	"$(MAGICK)" -background none -density 384 "$(SITE_LOGO)" \
		-resize 256x256 -gravity center -extent 256x256 \
		-define icon:auto-resize=256,128,64,48,32,16 "$(SITE_FAVICON_ICO)"
	cp "$(SITE_LOGO)" "$(SITE_FAVICON_SVG)"

.PHONY: site
site: kumbuka-cli ## Build the static documentation site with Kumbuka.
	"$(KUMBUKA_CLI)" build --config "$(SITE_CONFIG)"

.PHONY: site-serve
site-serve: kumbuka-cli ## Build and serve the documentation site locally.
	"$(KUMBUKA_CLI)" build --config "$(SITE_CONFIG)" \
		--site-url "http://127.0.0.1:$(SITE_PORT)/"
	@echo "Serving Kumbuka documentation at http://127.0.0.1:$(SITE_PORT)"
	python3 -m http.server $(SITE_PORT) --bind 127.0.0.1 --directory "$(SITE_OUTPUT)"

# Preserve the existing target name for callers.
.PHONY: tag
tag: current ## Show the current semantic version tag.

##@ Dependencies

.PHONY: kumbuka-cli
kumbuka-cli: $(GITHUB_RELEASE_INSTALL) ## Download the pinned Kumbuka binary into bin if necessary.
ifeq ($(KUMBUKA_CLI),$(LOCALBIN)/kumbuka-cli)
	$(call run-tool,$(GITHUB_RELEASE_INSTALL),\
		--repo kumbuka-me/cli \
		--tag "$(KUMBUKA_CLI_VERSION)" \
		--asset "kumbuka-cli_{version}_{os}_{arch}.tar.gz" \
		--binary kumbuka-cli \
		--target "$(KUMBUKA_CLI)")
else
	@command -v "$(KUMBUKA_CLI)" >/dev/null 2>&1 || { echo "Kumbuka binary not found: $(KUMBUKA_CLI)" >&2; exit 1; }
endif

.PHONY: golangci-lint
golangci-lint: $(GO_INSTALL_TOOL) ## Download the pinned golangci-lint locally if necessary.
	$(call run-tool,$(GO_INSTALL_TOOL),\
		--target "$(GOLANGCI_LINT)" \
		--package github.com/golangci/golangci-lint/v2/cmd/golangci-lint \
		--tool-version "$(GOLANGCI_LINT_VERSION)")
