
all:
	sh package_assets.sh
	go build

clean:
	go clean
	rm -rf templates
	rm -rf posts
	rm -rf assets
	rm -f brog_config.json
	rm -f brog.log
	rm -f brogger/base_assets.go

install: configure all
	go get

# Target to setup the build appropriately, use patched version of bin2go
configure: clean
	go get -u github.com/aybabtme/bin2go
