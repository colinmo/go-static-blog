# von Explaino Blog generator

This is a command line tool to integrate a git repository with a blog template system. On change in a git repository it calls this script, this script pulls down any changes since the last pull and generates HTML pages (blog posts etc.). It uses frontmatter to give metadata to the blog pages, as well as directory structure to set article types.

## Build

```sh
go mod init github.com/colinmo/vonblog
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/gernest/front
go get github.com/gomarkdown/markdown 
go get -u github.com/tyler-sommer/stick
go get -u github.com/tyler-sommer/stick/twig
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

### Stick

Golang conversion of `twig` template tool for PHP

Twig extra extensions adds macros familiar to twig authors

### Godog

Cucumber style Go testing

### SonarQube

* Server
  * [Local run server](https://docs.sonarqube.org/latest/setup/get-started-2-minutes/)
  * Docker: `docker run -d --name sonarqube -e SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true -p 9000:9000 sonarqube:latest`
* Client
  * [Local run client](https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/)
  * Docker: `docker run --rm -v "f:/dropbox/swap/golang/vonblog:/usr/src" sonarsource/sonar-scanner-cli`
* Coverage - https://community.sonarsource.com/t/sonargo-code-coverage-0/19473

## XCompiling

### On Windows

```
cmd
set GOOS=linux
set GOARCH=amd64
go build -ldflags '-w -s'
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