# golang-common [![CircleCI](https://circleci.com/gh/monacohq/golang-common/tree/main.svg?style=shield&circle-token=daf1da839b5c2715ecf6e86532718dd83c4e5ca1)](https://circleci.com/gh/monacohq/golang-common/tree/main) [![Coverage Status](https://coveralls.io/repos/github/monacohq/golang-common/badge.svg?t=cPxXZ8)](https://coveralls.io/github/monacohq/golang-common)

common golang packages used at crypto.com

## Current modules

| module                      | benchmarks | latest version |
|---|---|---|
| [transport/http/handlerwrap](transport/http/handlerwrap) | [benches](https://turbo-winner-7f9425af.pages.github.io/transport/http/handlerwrap/) |3.0.0|
| [transport/http/middleware/cryptouseruuid](transport/http/middleware/cryptouseruuid) | [benches](https://turbo-winner-7f9425af.pages.github.io/transport/http/middleware/cryptouseruuid) |1.0.1|
| [monitoring/otelinit](monitoring/otelinit) | [benches](https://turbo-winner-7f9425af.pages.github.io/monitoring/otelinit) |1.0.5|
| [config/secrets](config/secrets) | [benches](https://turbo-winner-7f9425af.pages.github.io/config/secrets) |1.0.4|
| [database/pginit](database/pginit) | [benches](https://turbo-winner-7f9425af.pages.github.io/database/pginit) |1.3.1|

## How to use any of these private modules

Force the use of ssh instead of https for git:

```bash
git config --global --add url."git@github.com:".insteadOf "https://github.com/"
```

Allow internal repositories under monacohq, simply add this line to your .zshrc or other, accordingly:

```bash
export GOPRIVATE="github.com/monacohq/*"
```

## How to check out any of these private modules from the CircleCI in your project

- Prepare a (machine) user account to have access permission both to your project and this repository.
- Go to the `Project Settings` in CircleCI and select the `SSH Keys` menu
- Under the `Checkout SSH Keys` section, click on `Add User Key` to add the (machine) user key
- Proceed with authorization

More details can be found in [CircleCI docs](https://circleci.com/docs/github-integration#controlling-access-via-a-machine-user).

## How to add a new module

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

- Create a pull request
- Wait for review
