#!/usr/bin/bash
name=studiopollinator-api

if [ ! -f ./init/$name.service ]; then
  echo "missing ./init/$name.service file"
  exit 1
fi

go build -o ./bin/$name ./cmd/$name

# bundle up the deployment files
./scripts/package.sh $name ./deployment

# send the deployment to the server
rsync -rlpcgovziP ./deployment/ $WEBUSER@studiopollinator.com:deployments/$name/

# install the deployment on the server
ssh -t $WEBUSER@studiopollinator.com "sudo -s bash deployments/$name/install.sh $name deployments/$name"

# clean up the local deployment files
rm -rf ./deployment