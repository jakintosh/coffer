#!/usr/bin/bash
if [[ -f .env/test  ]]; then
  eval $(cat .env/test) ./salary
else
  echo "no .env/test file"
fi
