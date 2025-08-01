build: client

client:
        go build -o ./bin/coffer ./cmd/coffer

install: client
        sudo cp ./bin/coffer /usr/local/bin/coffer
