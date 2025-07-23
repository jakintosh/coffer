build:
	go build -o ./bin/coffer ./cmd/coffer

run: build
	./bin/coffer
