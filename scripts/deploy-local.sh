#!/usr/bin/bash

name=studiopollinator-api

if [ ! -f ./init/$name.service ]; then
  echo "missing ./init/$name.service file"
  exit 1
fi

if [ ! -f ./env/local/$name.env ]; then
  echo "missing ./env/local/$name.env file"
  exit 1
fi

sudo systemctl stop $name.service

go build -o ./bin/$name ./cmd/$name

sudo mkdir -p /etc/systemd/system
sudo mkdir -p /usr/local/env
sudo mkdir -p /usr/local/bin

sudo cp ./init/$name.service  /etc/systemd/system/
sudo cp ./env/local/$name.env /usr/local/env/
sudo cp ./bin/$name           /usr/local/bin/

sudo systemctl daemon-reload
