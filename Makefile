build: client server

client:
	go build -o ./bin/coffer ./cmd/coffer

server:
	go build -o ./bin/coffer-server ./cmd/coffer-server

install: client
	sudo cp ./bin/coffer /usr/local/bin/coffer

run: build
	./bin/coffer
