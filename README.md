# golang-common [![Coverage Status](https://coveralls.io/repos/github/monacohq/golang-common/badge.svg?t=cPxXZ8)](https://coveralls.io/github/monacohq/golang-common)

common golang packages used at crypto.com

## Current modules

| module                      | version | audit | benchmarks |
|---|---|---|---|
| transport/http/handlewrap   | | [![audit](https://github.com/monacohq/golang-common/actions/workflows/transport_http_handlewrap.audit.yml/badge.svg)](https://github.com/monacohq/golang-common/actions/workflows/transport_http_handlewrap.audit.yml) | [benches](https://turbo-winner-7f9425af.pages.github.io/transport/http/handlerwrap/) |
| monitoring/otelinit         | | [![audit](https://github.com/monacohq/golang-common/actions/workflows/monitoring_otelinit.audit.yml/badge.svg)](https://github.com/monacohq/golang-common/actions/workflows/monitoring_otelinit.audit.yml) | [benches](https://turbo-winner-7f9425af.pages.github.io/monitoring/otelinit) |
| config/secrets         | | [![audit](https://github.com/monacohq/golang-common/actions/workflows/config_secrets.audit.yml/badge.svg)](https://github.com/monacohq/golang-common/actions/workflows/config_secrets.audit.yml) | [benches](https://turbo-winner-7f9425af.pages.github.io/config/secrets) |

## How to use any of these private modules

Force the use of ssh instead of https for git:

```bash
git config --global --add url."git@github.com:".insteadOf "https://github.com/"
```

Allow internal repositories under monacohq, simply add this line to your .zshrc or other, accordingly:

```bash
export GOPRIVATE="github.com/monacohq/*"
```

## How to add a new module?

Let's take an example of an opentelemetry module.

- Make sure the module is fully tested (at least 95% coverage, try to reach 100%), linted
- Create a branch feat/opentelemetry
- Copy in the right folder (that's quite subjective), in our case, ./monitoring/otelinit
- Add it to the workspace

```bash
    go work use ./monitoring/otelinit
```

- Add your file, commit your files (respecting conventional commits) and tag the commit properly, according to a semantic versioning

```bash
    git add ./monitoring/otelinit
    git commit -m "feat: add monitoring opentelemetry module" ./monitoring/otelinit
    git tag "monitoring/otelinit/v1.0.0"
```

**IMPORTANT** Note the folder and subfolders in the tag.

- Make sure to add your workflows inside the root .github/workflows directory, with a name that makes sense
  - beware of your action name
  - beware of your "on" triggers
  - beware of the necessary working_directories, related to your module

- Add your module to the coverage github workflow

- Create a pull request
- Wait for review
