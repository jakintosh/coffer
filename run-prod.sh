#!/usr/bin/bash
if [[ -f .env/prod  ]]; then
  eval $(cat .env/prod) ./salary
else
  echo "no .env/prod file"
fi
