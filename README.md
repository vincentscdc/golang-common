# golang-common

common golang packages used at crypto.com

## Current modules

transport/http/handlewrap
monitoring/otelinit

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
  
- Create a pull request
- Wait for review
