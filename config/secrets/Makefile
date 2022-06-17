.PHONY: help test test-leak bench bench-compare lint sec-scan upgrade release release-tag changelog-gen changelog-commit docs-gen

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


MOD_NAME="config/secrets"

########
# test #
########

test: ## launch tests for a specific module
	go test -v ./... -cover -covermode=atomic -coverprofile=./coverage.out

# test-leak: ## launch tests for a specific module with leak detection (if enabled)
# 	go test ./ -leak

bench: ## launch benchs
	go test ./... -bench=. -benchmem | tee ./bench.txt

bench-compare: ## compare benchs results
	benchstat ./bench.txt

########
# lint #
########

lint: ## lints the entire codebase
	@golangci-lint run ./... --config=./.golangci.toml && \
	if [ $$(gofumpt -e -l ./ | wc -l) == "0" ] ; \
		then exit 0; \
	else \
		echo "these files needs to be gofumpt-ed"; \
		gofumpt -e -l ./; \
		exit 1; \
	fi

#######
# sec #
#######

sec-scan: ## scan for sec issues with trivy (trivy binary needed)
	trivy fs ./

############
# upgrades #
############

upgrade: ## upgrade dependencies (beware, it can break everything)
	go mod tidy && \
	go get -t -u ./... && \
	go mod tidy

###########
# release #
###########

MOD_VERSION = $(shell git describe --abbrev=0 --tags --match "$(MOD_NAME)/*")

release: release-tag changelog-gen changelog-commit ## create a new tag to release this module
	
release-tag: 
	@printf "here is the latest tag present: "; \
	printf "$(MOD_VERSION)\n"; \
	printf "what tag do you want to give? (use the form $(MOD_NAME)/vX.X.X): "; \
	read -r TAG && \
	git tag $$TAG && \
	printf "\nrelease tagged $$TAG !\n"

#############
# changelog #
#############

MESSAGE_CHANGELOG_COMMIT="docs: update CHANGELOG.md for $(MOD_VERSION)"

changelog-gen: 
	@git cliff \
		--include-path "**/$(MOD_NAME)/*" \
		--repository ../../ \
		-o ./CHANGELOG.md && \
	printf "\nchangelog generated!\n"

# keep this commit unconventional so it doesnt appear in the changelog
changelog-commit:
	git commit -m $(MESSAGE_CHANGELOG_COMMIT) ./CHANGELOG.md

########
# docs #
########

docs-gen: ## generate docs as markdown thanks to https://github.com/princjef/gomarkdoc
	@gomarkdoc --output README.md . && \
	printf "\ndocs generated!\n"