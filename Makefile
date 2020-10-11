# Copyright 2012 tsuru authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

GOCMD=go
BUILD_DIR = build
C8Y_PKGS = $$(go list ./... | grep -v /vendor/)
GOMOD=$(GOCMD) mod
TEST_THROTTLE_LIMIT=10

# Set VERSION from git describe
VERSION := $(shell git describe | sed "s/^v\?\.\([0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}\).*/\1/")

ENV_FILE ?= c8y.env
-include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE) 2>/dev/null)

.PHONY: all check-path test race docs install tsurud

all: check-path build test

show-version:		## Show current version
	@echo "VERSION: $(VERSION)"

# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

# It does not support GOPATH with multiple paths.
check-path:
	ifndef GOPATH
		@echo "FATAL: you must declare GOPATH environment variable, for more"
		@echo "       details, please check"
		@echo "       http://golang.org/doc/code.html#GOPATH"
		@exit 1
	endif
	@exit 0

check-integration-variables:
	$(call check_defined, C8Y_HOST, Cumulocity host url. i.e. https://cumulocity.com)
	$(call check_defined, C8Y_TENANT , Cumulocity tenant)
	$(call check_defined, C8Y_USER, Cumulocity username)
	$(call check_defined, C8Y_PASSWORD, Cumulocity password)
	@exit 0

gh_pages_install:	## Install github pages dependencies for viewing docs locally
	cd docs && bundle install

gh_pages_update:	## Update github pages dependencies
	cd docs && bundle update

gh_pages:			## Run github pages locally
	cd docs && bundle exec jekyll server --baseurl ""

docs-powershell: build		## Update the powershell docs
	pwsh -File ./scripts/build-powershell/build-docs.ps1 -Recreate

test: test_powershell test_powershell_sessions

lint: metalint

install:
	go mod download

metalint:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	go install $(C8Y_PKGS)
	go test -i $(C8Y_PKGS)
	echo "$(C8Y_PKGS)" | sed 's|github.com/tsuru/tsuru/|./|' | xargs -t -n 4 \
		time golangci-lint run -c ./.golangci.yml

race:
	go test $(GO_EXTRAFLAGS) -race -i $(C8Y_PKGS)
	go test $(GO_EXTRAFLAGS) -race $(C8Y_PKGS)


docs:
	godoc -http=":6060"

release:
	@if [ ! $(version) ]; then \
		echo "version parameter is required... use: make release version=<value>"; \
		exit 1; \
	fi

	$(eval PATCH := $(shell echo $(version) | sed "s/^\([0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}\).*/\1/"))
	$(eval MINOR := $(shell echo $(PATCH) | sed "s/^\([0-9]\{1,\}\.[0-9]\{1,\}\).*/\1/"))
	@if [ $(MINOR) == $(PATCH) ]; then \
		echo "invalid version"; \
		exit 1; \
	fi

	@if [ ! -f docs/releases/go-c8y-cli/$(PATCH).rst ]; then \
		echo "to release the $(version) version you should create a release notes for version $(PATCH) first."; \
		exit 1; \
	fi

	@echo "Releasing go-c8y $(version) version."
	@echo "Replacing version string."

	# @git add docs/conf.py api/server.go
	@git commit -m "bump to $(version)"

	@echo "Creating $(version) tag."
	@git tag $(version)

	@git push --tags
	@git push origin master

	@echo "$(version) released!"

update-vendor:
	GO111MODULE=on $(GOMOD) download
	GO111MODULE=on $(GOMOD) vendor

cpu-trace:
	$(GOCMD) test -bench=. -cpuprofile cpu.prof github.com/reubenmiller/go-c8y-cli/pkg/cmd
	$(GOCMD) tool pprof -svg cpu.prof > cpu.svg

trace:
	GO111MODULE=on $(GOCMD) test -trace trace.out github.com/reubenmiller/go-c8y-cli/pkg/cmd

view-trace:
	$(GOCMD) tool trace trace.out



profile:
	GO111MODULE=on $(GOCMD) test -bench=Div_SSA -cpuprofile=cpu.pb.gz github.com/reubenmiller/go-c8y-cli/pkg/cmd

build: update_spec build_cli build_powershell

build_cli:
	pwsh -File scripts/build-cli/build.ps1;

#
# Powershell Module
#
update_spec:
	pwsh -File scripts/generate-spec.ps1;

build_powershell:
	pwsh -File scripts/build-powershell/build.ps1;

test_powershell:
	pwsh -NonInteractive -File tools/PSc8y/test.parallel.ps1 -ThrottleLimit $(TEST_THROTTLE_LIMIT) -TestFileExclude "Set-Session|Get-SessionHomePath"
	# pwsh -NonInteractive -File tools/PSc8y/tests.ps1

test_powershell_sessions:		## Run tests which interfere with the session variable
	pwsh -NonInteractive -File tools/PSc8y/test.parallel.ps1 -ThrottleLimit 1 -TestFileFilter "Set-Session|Get-SessionHomePath"

test_bash:
	./tools/bash/tests/test.sh

install_c8y: build			## Install c8y in dev environment
	@if [ ! -f /usr/local/bin/c8y ]; then \
		sudo ln -s "$$(pwd)/tools/PSc8y/Dependencies/c8y.linux" /usr/local/bin/c8y; \
	fi
	@cp ./tools/bash/c8y.profile.sh ~/
	@echo "source ~/c8y.profile.sh"  >> ~/.bashrc

	@echo Installed c8y successfully

publish:
	pwsh -File ./scripts/build-powershell/publish.ps1

build-docker:
	@cp tools/PSc8y/Dependencies/c8y.linux ./docker/c8y.linux
	@cp tools/bash/c8y.plugin.zsh ./docker/c8y.plugin.zsh
	@cp tools/bash/c8y.profile.sh ./docker/c8y.profile.sh

	@sudo docker build ./docker --file ./docker/zsh.dockerfile $(DOCKER_BUILD_ARGS) --build-arg C8Y_VERSION=$(VERSION) --tag $(TAG_PREFIX)c8y-zsh
	@sudo docker build ./docker --file ./docker/bash.dockerfile $(DOCKER_BUILD_ARGS) --build-arg C8Y_VERSION=$(VERSION) --tag $(TAG_PREFIX)c8y-bash
	@sudo docker build ./docker --file ./docker/pwsh.dockerfile $(DOCKER_BUILD_ARGS) --tag $(TAG_PREFIX)c8y-pwsh

	@rm ./docker/c8y.linux
	@rm ./docker/c8y.plugin.zsh
	@rm ./docker/c8y.profile.sh

publish-docker: show-version build build-docker		## Publish docker c8y cli images
	@chmod +x ./scripts/publish-docker.sh
	@sudo CR_PAT=$(CR_PAT) VERSION=$(VERSION) ./scripts/publish-docker.sh

run-docker-bash:
	sudo docker run -it --rm \
		-e C8Y_USE_ENVIRONMENT=true \
		-e C8Y_HOST=$$C8Y_HOST \
		-e C8Y_TENANT=$$C8Y_TENANT \
		-e C8Y_USER=$$C8Y_USER \
		-e C8Y_PASSWORD=$$C8Y_PASSWORD \
		c8y-bash

run-docker-zsh:
	sudo docker run -it --rm \
		-e C8Y_USE_ENVIRONMENT=true \
		-e C8Y_HOST=$$C8Y_HOST \
		-e C8Y_TENANT=$$C8Y_TENANT \
		-e C8Y_USER=$$C8Y_USER \
		-e C8Y_PASSWORD=$$C8Y_PASSWORD \
		c8y-zsh

run-docker-pwsh:
	sudo docker run -it --rm \
		-e C8Y_USE_ENVIRONMENT=true \
		-e C8Y_HOST=$$C8Y_HOST \
		-e C8Y_TENANT=$$C8Y_TENANT \
		-e C8Y_USER=$$C8Y_USER \
		-e C8Y_PASSWORD=$$C8Y_PASSWORD \
		c8y-pwsh
