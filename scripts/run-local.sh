#!/usr/bin/bash

name=studiopollinator-api

if [ -f env/local/$name.env ]; then
  source env/local/$name.env
else
  echo "missing env/local/$name.env file"
  exit 1
fi

sudo systemctl stop $name.service
sudo systemctl start $name.service
stripe listen --forward-to localhost:$PORT/webhook
