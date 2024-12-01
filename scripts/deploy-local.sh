#!/usr/bin/bash
name=studiopollinator-api

if [ ! -f ./init/$name.service ]; then
  echo "missing ./init/$name.service file"
  exit 1
fi

go build -o ./bin/$name ./cmd/$name

# bundle up the deployment
./scripts/package.sh $name ./deployment

# install the deployment files
./scripts/install.sh $name ./deployment

# clean up deployment files
rm -r ./deployment
