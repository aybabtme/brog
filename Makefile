
all:
	go build

clean:
	go clean
	rm -rf templates
	rm -rf posts
	rm -rf assets
	rm -f brog_config.json
	rm -f brog.log

install: clean
	go get

# Target to setup the build appropriately
configure: clean
	go get github.com/chsc/bin2go
