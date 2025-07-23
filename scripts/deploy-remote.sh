#!/usr/bin/bash
DOMAIN=${1:?"Domain name required."} || exit 1

name=coffer
dpl_src=./deployment
dpl_dst=deployments

if [ ! -f ./init/$name.service ]; then
  echo "missing ./init/$name.service file"
  exit 1
fi

# build the executable
./scripts/build.sh

# bundle up the deployment files
./scripts/package.sh $name $dpl_src

# send the deployment to the server
rsync -a --del $dpl_src/ $WEBUSER@$DOMAIN:$dpl_dst/$name/

# install the deployment on the server
ssh -t $WEBUSER@$DOMAIN "sudo -s bash $dpl_dst/$name/install.sh $name $DOMAIN $dpl_dst/$name"

# clean up the local deployment files
rm -r $dpl_src
