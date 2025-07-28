build: client server

client:
	go build -o ./bin/coffer ./cmd/coffer

server:
	go build -o ./bin/coffer-server ./cmd/coffer-server

run: build
	./bin/coffer
