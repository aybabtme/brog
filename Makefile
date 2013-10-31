
all:
	go build

clean:
	go clean
	rm -rf templates
	rm -rf posts
	rm -f brog_config.json
	rm -f brog.log

install: clean
	go get
