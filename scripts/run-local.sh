#!/usr/bin/bash
name=studiopollinator-api

sudo systemctl stop $name.service
sudo systemctl start $name.service
stripe listen --forward-to localhost:9000/webhook
