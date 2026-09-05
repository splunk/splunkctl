.PHONY: build build-release build-runner test generate-integration generate-integration-local test-integration test-integration-local fmt fmt-check generate clean

include VERSION

APP_NAME     := splunkctl
BUILD_PATH   := $(or $(CI_PROJECT_DIR),$(CURDIR))/bin
RELEASE_BUILD_PATH := $(or $(CI_PROJECT_DIR),$(CURDIR))/artifacts/github-release/build
OS       := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH     := $(shell uname -m)
PLATFORM := $(OS)_$(ARCH)


INTEGRATION_MODE ?= test
INTEGRATION_CMD ?=
SPLUNKCTL_BIN ?= $(CURDIR)/bin/splunkctl
INTEGRATION_RUNNER ?= $(CURDIR)/bin/integration-runner
LOCAL_DEPLOYMENT_FILE ?= cicd/tools/orca/deployment/local-deployment.yml

build-runner: build
	go build -o bin/integration-runner ./tests/integration/runner

build:
	$(eval COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none))
	$(eval BUILT  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	mkdir -p $(BUILD_PATH)
	go build -ldflags "-X github.com/splunk/splunkctl/cmd.Version=$(version) -X github.com/splunk/splunkctl/cmd.Commit=$(COMMIT) -X github.com/splunk/splunkctl/cmd.Built=$(BUILT)" -o $(BUILD_PATH)/$(APP_NAME) .

build-release:
	$(eval COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none))
	$(eval BUILT  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	mkdir -p $(RELEASE_BUILD_PATH)/linux_amd64 $(RELEASE_BUILD_PATH)/linux_arm64 $(RELEASE_BUILD_PATH)/darwin_amd64 $(RELEASE_BUILD_PATH)/darwin_arm64 $(RELEASE_BUILD_PATH)/windows_amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X github.com/splunk/splunkctl/cmd.Version=$(version) -X github.com/splunk/splunkctl/cmd.Commit=$(COMMIT) -X github.com/splunk/splunkctl/cmd.Built=$(BUILT)" -o $(RELEASE_BUILD_PATH)/linux_amd64/$(APP_NAME) .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-X github.com/splunk/splunkctl/cmd.Version=$(version) -X github.com/splunk/splunkctl/cmd.Commit=$(COMMIT) -X github.com/splunk/splunkctl/cmd.Built=$(BUILT)" -o $(RELEASE_BUILD_PATH)/linux_arm64/$(APP_NAME) .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X github.com/splunk/splunkctl/cmd.Version=$(version) -X github.com/splunk/splunkctl/cmd.Commit=$(COMMIT) -X github.com/splunk/splunkctl/cmd.Built=$(BUILT)" -o $(RELEASE_BUILD_PATH)/darwin_amd64/$(APP_NAME) .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-X github.com/splunk/splunkctl/cmd.Version=$(version) -X github.com/splunk/splunkctl/cmd.Commit=$(COMMIT) -X github.com/splunk/splunkctl/cmd.Built=$(BUILT)" -o $(RELEASE_BUILD_PATH)/darwin_arm64/$(APP_NAME) .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X github.com/splunk/splunkctl/cmd.Version=$(version) -X github.com/splunk/splunkctl/cmd.Commit=$(COMMIT) -X github.com/splunk/splunkctl/cmd.Built=$(BUILT)" -o $(RELEASE_BUILD_PATH)/windows_amd64/$(APP_NAME).exe .

package: build
	cp VERSION $(BUILD_PATH)/VERSION
	cp LICENSE $(BUILD_PATH)/LICENSE
	cd $(BUILD_PATH) && tar -czf $(APP_NAME)_$(version)_$(COMMIT).tar.gz $(APP_NAME) VERSION LICENSE

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-integration:
	@test -x "$(SPLUNKCTL_BIN)" || { echo "Missing $(SPLUNKCTL_BIN), run 'make build-runner' to create binaries and run int tests"; exit 1; }
	@test -x "$(INTEGRATION_RUNNER)" || { echo "Missing $(INTEGRATION_RUNNER)"; exit 1; }
	"$(INTEGRATION_RUNNER)" \
		--mode "$(INTEGRATION_MODE)" \
		--cmd "$(INTEGRATION_CMD)" \
		--binary "$(SPLUNKCTL_BIN)"

test-integration-local: build-runner
	@set -e; \
	config_file="$$(mktemp)"; \
	trap 'rm -f "$$config_file"' EXIT; \
	python3 cicd/tools/orca/orca_service.py configure-local \
		--deployment-file "$(LOCAL_DEPLOYMENT_FILE)" \
		--config-file "$$config_file"; \
	export SPLUNKCTL_HOST="$$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1]))["host"])' "$$config_file")"; \
	export SPLUNKCTL_TOKEN="$$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1]))["token"])' "$$config_file")"; \
	$(MAKE) test-integration \
		INTEGRATION_MODE="$(INTEGRATION_MODE)" \
		INTEGRATION_CMD="$(INTEGRATION_CMD)"

generate-integration: INTEGRATION_MODE=generate test-integration

fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following Go files are not formatted:"; \
		printf '%s\n' "$$unformatted"; \
		exit 1; \
	fi

fmt:
	go fmt ./...

generate:
	go generate ./...

clean:
	rm -rf $(BUILD_PATH)
