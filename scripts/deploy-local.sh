#!/usr/bin/bash

if [ ! -f env/local/studiopollinator-api.env ]; then
  echo "no env/local/studiopollinator-api.env file"
  exit 1
fi

sudo systemctl stop studiopollinator-api.service

go build -o bin/studiopollinator-api ./cmd/studiopollinator-api

sudo mkdir -p /etc/systemd/system
sudo mkdir -p /usr/local/bin
sudo mkdir -p /usr/local/env

sudo cp ./init/studiopollinator-api.service  /etc/systemd/system/
sudo cp ./bin/studiopollinator-api           /usr/local/bin/
sudo cp ./env/local/studiopollinator-api.env /usr/local/env/

sudo systemctl daemon-reload
