.PHONY: docs-gen changelog-gen version-table-update

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: docs-gen changelogs-gen version-table-update

ALL_MODULES=$(shell git tag --merged main | sed -E 's:/v[0-9]+.*::' | uniq | tr '\n' ' ')

########
# docs #
########

docs-gen: ## generate docs for every module, as markdown thanks to https://github.com/princjef/gomarkdoc
	@( \
		for module in $(ALL_MODULES); do \
			gomarkdoc --output ./$$module/README.md ./$$module/; \
			printf "docs generated for $$module!\n"; \
			git commit -m "docs: update docs for module $$module for tag $(shell git describe --abbrev=0 --tags --match "$$module/*")" ./$$module/README.md; \
		done \
	)

##############
# changelogs #
##############

changelogs-gen: ## Generate changelog for every module.
	@( \
		for module in $(ALL_MODULES); do \
			git cliff \
				--include-path "**/$$module/*" \
				-c ./$$module/cliff.toml \
				-o ./$$module/CHANGELOG.md; \
			printf "\nchangelog generated for $$module!\n"; \
			git commit -m "docs(changelog): update CHANGELOG.md for $(shell git describe --abbrev=0 --tags --match "$$module/*")" ./$$module/CHANGELOG.md; \
		done \
	)


#############################
# versions update in README #
#############################

version-table-update: ## Update version table in README.md to latest version.
	@( \
		git tag | sed -E 's:/v[0-9]+.*::' | uniq | xargs -S 512 -I{} sh -c \
		'VERSION=$$(git tag --list "{}/v*" | sed "s:{}/v::" | sort -t . -k1n -k2n -k3n | tail -n 1);\
		sed -i "" "\:{}.*benches:  s:|\(.*\)|\(.*\)|\(.*\)|:|\1|\2|$$VERSION|:" README.md';\
		git commit -m "docs(readme): update latest versions" ./README.md; \
	)

########
# lint #
########

lint: ## lints the entire codebase
	@golangci-lint run $(ALL_MODULES) && \
	if [ $$(gofumpt -e -l ./ | wc -l) == "0" ] ; then exit 0; else exit 1; fi

#######
# sec #
#######

sec-scan: ## scan for sec issues with trivy (trivy binary needed)
	trivy fs ./

#########
# tests #
#########

ALL_MODULES_DOTDOTDOT = $(shell git tag --merged main | sed -E 's:/v[0-9]+.*::' | uniq | xargs printf "./%s/... ")

test: ## launch tests for all modules
	go test -v $(ALL_MODULES_DOTDOTDOT) 

coveralls: ## launch tests for all modules and send them to coveralls
	@( 
		go test -v $(ALL_MODULES_DOTDOTDOT) -covermode=atomic -coverprofile=./coverage.out; \
		goveralls -covermode=atomic -coverprofile=./coverage.out -service=circle-ci -repotoken=$(COVERALLS_TOKEN) \
	)
