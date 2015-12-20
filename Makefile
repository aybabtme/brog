
# Packages assets into go source code, do a build and do static analysis
all:
	@echo "> Preparing a build."
	sh package_assets.sh
	go build
	sh findbugs.sh
	go test ./...

# Remove the junk that's generated by a build
clean:
	@echo "> Cleaning up development artifacts."
	go clean
	rm -rf templates
	rm -rf posts
	rm -rf pages
	rm -rf assets
	rm -f brog_config.json
	rm -f brog.log

# Sets up a system to be able to build brog
configure: clean
	@echo "> Preparing system for build."
	ln -s ../../githooks/pre-commit .git/hooks/pre-commit
	go get -u github.com/chsc/bin2go
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/vet
	go get -u github.com/golang/lint/golint
	go get -t ./...

# Perform all the steps to install a clean build of brog
install: configure all
	@echo "> Installing brog."
	go get -t ./...
