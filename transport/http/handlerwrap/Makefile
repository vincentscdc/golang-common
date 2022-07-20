.PHONY: help test test-leak bench bench-compare lint sec-scan upgrade release release-tag changelog-gen changelog-commit docs-gen

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


MOD_NAME="transport/http/handlerwrap"

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