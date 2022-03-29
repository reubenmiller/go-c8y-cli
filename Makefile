#
# Variables
#
TEST_THROTTLE_LIMIT=10
TEST_FILE_FILTER = .+
GITHUB_TOKEN ?=

# Set VERSION from git describe
VERSION := $(shell git describe | sed "s/^v\?\([0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}\).*/\1/")

ENV_FILE ?= c8y.env
-include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE) 2>/dev/null)
export C8Y_SETTINGS_CI=true

.PHONY: all test install docs-c8y manpages

all: build test

# ---------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------
show-version:		## Show current version
	@echo "VERSION: $(VERSION)"

init_setup: install_c8y
	pwsh -File ./scripts/build-powershell/install.ps1

install:
	go mod download

install_c8y: build			## Install c8y in dev environment
	@if [ ! -f /usr/local/bin/c8y ]; then \
		sudo ln -s "$$(pwd)/tools/PSc8y/Dependencies/c8y.linux" /usr/local/bin/c8y; \
	fi
	@cp ./tools/shell/c8y.plugin.sh ~/
	@echo "source ~/c8y.plugin.sh"  >> ~/.bashrc

	@echo Installed c8y successfully

# ---------------------------------------------------------------
# Docs
# ---------------------------------------------------------------
docs: docs-powershell docs-c8y		## Update all docs

docs-powershell: build		## Update the powershell docs
	pwsh -File ./scripts/build-powershell/build-docs.ps1 -Recreate -OutputFolder "docs/go-c8y-cli/docs/cli/psc8y"

docs-c8y:					## create c8y documentation
	go run ./cmd/gen-docs --website --doc-path "docs/go-c8y-cli/docs/cli/c8y"

manpages:					## create c8y man packages
	go run ./cmd/gen-docs --man-page --doc-path "./share/man/man1/"


# ---------------------------------------------------------------
# Github pages
# ---------------------------------------------------------------
gh_pages_install:	## Install github pages dependencies for viewing docs locally
	cd docs/go-c8y-cli && npm install

gh_pages:			## Run github pages locally
	cd docs/go-c8y-cli && npm start

build_gh_pages:		## build github pages
	cd docs/go-c8y-cli \
		&& npm install \
		&& npm run write-translations \
		&& npm run build

gh_pages_prod:  build_gh_pages		## build/serve github pages (production)
	cd docs/go-c8y-cli \
		&& npm run serve

# ---------------------------------------------------------------
# Spec and code generation
# ---------------------------------------------------------------
update_spec:					## Update json specifications
	pwsh -File scripts/generate-spec.ps1;

generate_go_code: update_spec		## Generate go code from spec
	pwsh -File scripts/build-cli/build.ps1 -SkipBuildBinary;

generate_cli_tests:				## Generate CLI tests from the examples in the api specs
	chmod +x ./scripts/build-cli-tests.sh
	./scripts/build-cli-tests.sh

# ---------------------------------------------------------------
# Linting
# ---------------------------------------------------------------
lint:
	golangci-lint run

# ---------------------------------------------------------------
# Build
# ---------------------------------------------------------------
build: update_spec build_cli build_powershell generate_cli_tests

build_cli:						## Generate the cli code and build the binaries
	pwsh -File scripts/build-cli/build.ps1;

build_cli_fast:					## Only build the linux version of the c8y binary
	pwsh -File ./scripts/build-cli/build-binary.ps1 -OutputDir ./tools/PSc8y/dist/PSc8y/Dependencies -Target "linux:$(dpkg --print-architecture)"
	cp ./tools/PSc8y/dist/PSc8y/Dependencies/c8y.linux /workspaces/go-c8y-cli/tools/PSc8y/Dependencies/c8y.linux

build_powershell:				## Build the powershell module
	pwsh -File scripts/build-powershell/build.ps1;

build-docker:					## Build the docker images
	@cp tools/PSc8y/Dependencies/c8y.linux ./docker/c8y.linux
	@sudo docker build ./docker --file ./docker/shell.dockerfile $(DOCKER_BUILD_ARGS) --build-arg C8Y_VERSION=$(VERSION) --tag $(TAG_PREFIX)c8y-shell
	@sudo docker build ./docker --file ./docker/pwsh.dockerfile $(DOCKER_BUILD_ARGS) --tag $(TAG_PREFIX)c8y-pwsh
	@rm ./docker/c8y.linux

build_completions:
	@chmod +x tools/shell/completion/build_completions.sh
	env OUTPUT_DIR=./output ./tools/shell/completion/build_completions.sh


checkout_addons:				## Checkout the go-c8y-cli-addons repository
	@chmod +x tools/shell/addons/checkout_addons.sh
	tools/shell/addons/checkout_addons.sh

# ---------------------------------------------------------------
# Tests
# ---------------------------------------------------------------
test: test_powershell test_powershell_sessions		## Run all tests

test_powershell:				## Run powershell tests
	pwsh -ExecutionPolicy bypass -NonInteractive -File tools/PSc8y/test.parallel.ps1 -ThrottleLimit $(TEST_THROTTLE_LIMIT) -TestFileFilter "$(TEST_FILE_FILTER)" -TestFileExclude "Set-Session|Get-SessionHomePath|Login|DisableCommands|BulkOperation|activitylog|Invoke-UserLogout"

test_powershell_sessions:		## Run powershell tests which interfere with the session variable
	pwsh -ExecutionPolicy bypass -NonInteractive -File tools/PSc8y/test.parallel.ps1 -ThrottleLimit 1 -TestFileFilter "Set-Session|Get-SessionHomePath|Login|DisableCommands|BulkOperation|activitylog|Invoke-UserLogout"

test_cli: test_cli_auto test_cli_manual

test_cli_dependencies:										## Install cli dependencies
	command -v commander || go get github.com/commander-cli/commander/v2/cmd/commander@v2.5.0-alpha3

test_cli_auto: test_cli_dependencies						## Run tests on the cli examples
	./tests/scripts/setup.sh
	./tests/run-auto.sh

test_cli_manual: test_cli_dependencies						## Run tests on the cli examples
	./tests/run-manual.sh

test_installation:				## Test installation of linux packages
	chmod +x ./tools/integration-tests/run.sh
	sudo ./tools/integration-tests/run.sh

# ---------------------------------------------------------------
# Publish
# ---------------------------------------------------------------
publish:					## Publish powershell module
	pwsh -File ./scripts/build-powershell/publish.ps1

publish-docker: show-version build build-docker		## Publish docker c8y cli images
	@chmod +x ./scripts/publish-docker.sh
	@sudo CR_PAT=$(CR_PAT) VERSION=$(VERSION) ./scripts/publish-docker.sh

publish-local-snapshot:		## Publish local snapshot release
	goreleaser --snapshot --skip-publish --rm-dist

publish-release:			## Publish release
	goreleaser --rm-dist

publish-linux-packages:		## Publish linux packages (apk,deb,rpm)
	@chmod +x ./tools/publish-packages.sh
	./tools/publish-packages.sh ./dist

# ---------------------------------------------------------------
# Docker examples
# ---------------------------------------------------------------
run-docker-shell:
	sudo docker run -it --rm \
		-e C8Y_HOST=$$C8Y_HOST \
		-e C8Y_TENANT=$$C8Y_TENANT \
		-e C8Y_USER=$$C8Y_USER \
		-e C8Y_PASSWORD=$$C8Y_PASSWORD \
		c8y-shell

run-docker-pwsh:
	sudo docker run -it --rm \
		-e C8Y_HOST=$$C8Y_HOST \
		-e C8Y_TENANT=$$C8Y_TENANT \
		-e C8Y_USER=$$C8Y_USER \
		-e C8Y_PASSWORD=$$C8Y_PASSWORD \
		c8y-pwsh
