build:
	go build -o ./bin/coffer ./cmd/coffer

install: build
	sudo cp ./bin/coffer /usr/local/bin/coffer
