build: build-windows-windows build-linux-windows

# Start in vonblog
# cd src
# go test "./..." -coverprofile="coverage.out" > junit.xml
# .\gosonar.exe --basedir c:/users/relap/dropbox/swap/golang/vonblog/src/cmd/ --coverage coverage.out --junit junit.xml
# cd ..
# docker run --rm -v "${PWD}:/usr/src" sonarsource/sonar-scanner-cli

test:
	cd c:/users/relap/dropbox\swap\golang\vonblog\src && go test "./..." -coverprofile="coverage.out" -v 2>&1 | go-junit-report > junit.xml
	cd c:/users/relap/dropbox\swap\golang\vonblog\src && gosonar --basedir c:/users/relap/dropbox\swap\golang\vonblog\src\cmd\ --coverage coverage.out --junit junit.xml

build-windows-windows:
	set GOOS=windows&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s -H=windowsgui" -o ../bin/vonblog.exe
	

build-linux-windows:
	docker run --rm -v $PWD/src:/usr/src/myapp -v $PWD/bin:/tmp/bin -w /usr/src/myapp golang:1.23.0-bullseye go build -ldflags "-w -s" -o /tmp/bin/vonblog
#	set GOOS=linux&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ../bin/vonblog

build-linux-mac:
	cd src && env GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o ../bin/vonblog

build-linux-linux: 
	cd src && GO_ENABLED=0 go build -ldflags "-w -s" -o ../bin/vonblog

sonar: test
	cd c:/users/relap/dropbox\swap\golang\vonblog && docker run --rm -v c:/users/relap/dropbox\swap\golang\vonblog:/usr/src sonarsource/sonar-scanner-cli

sonarqube:
	docker run -d --name sonarqube -e SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true -p 9000:9000 sonarqube:latest

clean:
	cd c:/users/relap/dropbox\swap\golang\vonblog\src && del vonblog
	cd c:/users/relap/dropbox\swap\golang\vonblog\src && del vonblog.exe
	cd c:/users/relap/dropbox\swap\golang\vonblog\src && del coverage.out
	cd c:/users/relap/dropbox\swap\golang\vonblog\src && del junit.xml
