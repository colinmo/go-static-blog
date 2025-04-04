# von Explaino Blog generator

This is a command line tool to integrate a git repository with a blog template system. On change in a git repository it calls this script, this script pulls down any changes since the last pull and generates HTML pages (blog posts etc.). It uses frontmatter to give metadata to the blog pages, as well as directory structure to set article types.

# TASKS

* [ ] Script to convert Golang test output to Sonarqube
* [x] Pull out HTML access component so it can be passed through as a parameter for mocking/ testing
* [x] Refactoring the codebase (buzzwords)
* [x] More unit tests for resiliance
* [x] Change the code-window image holders so they look less pillowy
* [ ] Add a Webmention extension
  * [ ] Verifies incoming webmentions
  * [ ] Saves it to a .js file specific to the file ID
* [x] Fix RSS feeds to not include drafts

## Build

```sh
go mod init github.com/colinmo/vonblog
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/gernest/front
go get github.com/gomarkdown/markdown 
go get github.com/hashicorp/go-memdb
go get github.com/cucumber/gherkin-go  
go get github.com/cucumber/godog
```

### Cobra and Viper

Command line tool maker

### Front

JSON or YAML frontmatter parser

### Markdown

Converts markdown into HTML

### Godog

Cucumber style Go testing

### SonarQube

* Server
  * [Local run server](https://docs.sonarqube.org/latest/setup/get-started-2-minutes/)
  * Docker: `docker run -d --name sonarqube -e SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true -p 9000:9000 -v../sonarqubedata:/opt/sonarqube/data docker.io/library/sonarqube:latest `
* Client
  * [Local run client](https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/)
  * Docker: `docker run --rm -v "%cd%:/usr/src" docker.io/library/sonarsource/sonar-scanner-cli`
  * Docker: `docker run --rm -v ${PWD}:/usr/src docker.io/library/sonarsource/sonar-scanner-cli`
* Coverage - https://community.sonarsource.com/t/sonargo-code-coverage-0/19473
  * `go test "./..." -coverprofile="coverage.out"`
  * `go test "./..." -coverprofile="coverage.out" -json > test-report.json`
  * `go test "./..." -coverprofile="coverage.out" -v 2>&1 | go-junit-report > junit.xml ; gosonar --basedir ${PWD}\src\cmd\ --coverage coverage.out --junit junit.xml`

## XCompiling

### On Windows

```
cmd
cd src
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-w -s"
``` 

### On Unix

```
cd src
export GOOS=linux
export GOARCH=amd64
go build -ldflags '-w -s'
``` 

### On Mac

```
env GOOS=darwin GOARCH=386 go build
```

## TODO

- [x] Write the most recent post snippet into an HTML page for embedding
- [x] Write the last X days activity to an SVG image.