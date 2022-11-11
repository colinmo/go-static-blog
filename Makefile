build: build-windows-windows build-linux-windows

test:
	cd f:\Dropbox\swap\golang\vonblog\src && go test "./..." -coverprofile="coverage.out" -v 2>&1 | go-junit-report > junit.xml
	cd f:\Dropbox\swap\golang\vonblog\src && gosonar --basedir f:\Dropbox\swap\golang\vonblog\src\cmd\ --coverage coverage.out --junit junit.xml

build-windows-windows:
	set GOOS=windows&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ../bin/vonblog.exe
	
build-linux-windows:
	set GOOS=linux&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ../bin/vonblog

sonar: test
	cd f:\Dropbox\swap\golang\vonblog && docker run --rm -v f:\Dropbox\swap\golang\vonblog:/usr/src sonarsource/sonar-scanner-cli

sonarqube:
	docker run -d --name sonarqube -e SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true -p 9000:9000 sonarqube:latest

clean:
	cd f:\Dropbox\swap\golang\vonblog\src && del vonblog
	cd f:\Dropbox\swap\golang\vonblog\src && del vonblog.exe
	cd f:\Dropbox\swap\golang\vonblog\src && del coverage.out
	cd f:\Dropbox\swap\golang\vonblog\src && del junit.xml
