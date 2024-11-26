#!/usr/bin/bash
if [ -f env/local/studiopollinator-api.env ]; then
  source env/local/studiopollinator-api.env
else
  echo "no env/local/studiopollinator-api.env file"
  exit 1
fi

sudo systemctl stop studiopollinator-api.service
sudo systemctl start studiopollinator-api.service
stripe listen --forward-to localhost:$PORT/webhook
