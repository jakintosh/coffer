#!/usr/bin/bash
name=studiopollinator-api

if [ ! -f ./init/$name.service ]; then
  echo "missing ./init/$name.service file"
  exit 1
fi

sudo systemctl stop $name.service

go build -o ./bin/$name ./cmd/$name

sudo mkdir -p /etc/$name

sudo cp    ./bin/$name  /usr/local/bin/
sudo cp -r ./init/.     /etc/systemd/system/
sudo cp -r ./secrets/.  /etc/$name/

sudo systemctl daemon-reload
