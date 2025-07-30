#!/usr/bin/bash
name=coffer-server
domain=localhost

sudo systemctl stop $name@$domain
sudo systemctl start $name@$domain
stripe listen --forward-to 127.0.0.1:9000/api/v1/stripe/webhook
