.PHONY: docs-gen changelogs-gen version-table-update lint sec-scan upgrade release test coveralls

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: docs-gen changelogs-gen version-table-update

ALL_MODULES=$(shell go work edit -json | grep ModPath | sed -E 's:^.*golang-common/(.*)":\1:' | sed -E 's:/v[0-9]+$$::')

ALL_MODULES_SPACE_SEP=$(shell echo $(ALL_MODULES) | xargs printf "%s ")

ALL_MODULES_DOTDOTDOT=$(shell echo $(ALL_MODULES) | xargs printf "./%s/... ")

########
# docs #
########

docs-gen: ## generate docs for every module, as markdown thanks to https://github.com/princjef/gomarkdoc
	@( \
		for module in $(ALL_MODULES_SPACE_SEP); do \
			gomarkdoc --output ./$$module/README.md ./$$module/ && \
			printf "docs generated for $$module!\n"; \
			git commit -m "docs: update docs for module $$module" ./$$module/README.md; \
		done \
	)


##############
# changelogs #
##############

changelogs-gen: ## Generate changelog for every module.
	@( \
		for module in $(ALL_MODULES_SPACE_SEP); do \
			sed -i '' -E "s:TAG_MODULE:$$module:g" ./cliff.toml && \
			git cliff \
				--include-path "**/$$module/*" \
				-o ./$$module/CHANGELOG.md && \
			sed -i '' -E "s:$$module:TAG_MODULE:g" ./cliff.toml && \
			printf "\nchangelog generated for $$module!\n"; \
			git commit -m "docs(changelog): update CHANGELOG.md for $$(git describe --abbrev=0 --tags --match "$$module/*")" ./$$module/CHANGELOG.md; \
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
	@golangci-lint run $(ALL_MODULES_SPACE_SEP) && \
	if [ $$(gofumpt -e -l ./ | wc -l) = "0" ] ; \
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


###########
# upgrade #
###########

upgrade: ## upgrade selection module dependencies (beware, it can break everything)
	@( \
		select module in $(ALL_MODULES_SPACE_SEP); do \
			if [ -z $$module ]; then \
				break; \
			fi; \
			pushd $$module > /dev/null && \
			go mod tidy && \
			go get -t -u ./... && \
			go mod tidy && \
			popd > /dev/null; \
			break; \
		done \
	)


###########
# release #
###########

release: ## release selection module, gen-changelog, commit and tag
	@( \
		select module in $(ALL_MODULES_SPACE_SEP); do \
			if [ -z $$module ]; then \
				break; \
			fi; \
			printf "here is the $$module latest tag present: "; \
			git describe --abbrev=0 --tags --match "$$module/*"; \
			printf "what tag do you want to give? (use the form $$module/vX.X.X): "; \
			read -r TAG; \
			sed -i '' -E "s:TAG_MODULE:$$module:g" ./cliff.toml && \
			git cliff \
				--tag $$TAG \
				--include-path "**/$$module/*" \
				-o ./$$module/CHANGELOG.md && \
			sed -i '' -E "s:$$module:TAG_MODULE:g" ./cliff.toml && \
			printf "\nchangelog generated for $$module!\n"; \
			git commit -m "docs(changelog): update CHANGELOG.md for $$TAG" ./$$module/CHANGELOG.md && \
			git tag $$TAG && \
			printf "\nrelease tagged $$TAG !\n"; \
			printf "\nrelease and tagging has been done, if you are OK with everything, just git push origin $$(git describe --abbrev=0 --tags --match "$$module/*")\n"; \
			break; \
		done \
	)


#########
# tests #
#########

test: ## launch tests for all modules
	go test -v $(ALL_MODULES_DOTDOTDOT)

coveralls: ## launch tests for all modules and send them to coveralls
	@( \
		go test -v $(ALL_MODULES_DOTDOTDOT) -covermode=atomic -coverprofile=./coverage.out; \
		goveralls -covermode=atomic -coverprofile=./coverage.out -service=circle-ci -repotoken=$(COVERALLS_TOKEN) \
	)
