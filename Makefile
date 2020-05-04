# Copyright 2012 tsuru authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

GOCMD=go
BUILD_DIR = build
C8Y_PKGS = $$(go list ./... | grep -v /vendor/)
GOMOD=$(GOCMD) mod

.PHONY: all check-path test race docs install tsurud

all: check-path build test

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


test: test_powershell

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
	pwsh -File tools/PSc8y/tests.ps1 -NonInteractive

test_ci_powershell:
	pwsh -File scripts/build-powershell/test.ci.ps1 -NonInteractive
