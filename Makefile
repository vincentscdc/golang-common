.PHONY: docs-gen changelog-gen version-table-update

all: docs-gen changelog-gen version-table-update

##############
# Management #
##############

# Generate docs for every module.
#
# For every git tag, strip the version suffix (/v...) to get the module path.
docs-gen:
	@git tag | sed -E 's:/v[0-9]+.*::' | uniq | xargs -I{} sh -c "echo generate docs for {}; cd {}; make docs-gen; echo"


# Generate changelog for every module.
changelog-gen:
	@git tag | sed -E 's:/v[0-9]+.*::' | uniq | xargs -I{} sh -c "echo generate changelog for {}; cd {}; make changelog-gen; echo"


# Update version table in README.md to latest version.
#
# The sed command first matches the line containing the module and benchmark link,
# then substitutes the last column with the latest version.
version-table-update:
	@echo update version table
	@git tag | sed -E 's:/v[0-9]+.*::' | uniq | xargs -S 512 -I{} sh -c \
		'VERSION=$$(git tag --list "{}/v*" | sed "s:{}/v::" | sort -t . -k1n -k2n -k3n | tail -n 1);\
		 sed -i "" "\:{}.*benches:  s:|\(.*\)|\(.*\)|\(.*\)|:|\1|\2|$$VERSION|:" README.md'
