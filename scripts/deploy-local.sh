#!/usr/bin/bash
name=coffer-server
domain=localhost
dpl_src=./deployment

if [ ! -f ./init/$name@.service ]; then
  echo "missing ./init/$name@.service file"
  exit 1
fi

# build the executable
make

# bundle up the deployment files
./scripts/package.sh $name $dpl_src

# install the deployment files
$dpl_src/install.sh "$name" "$domain" "$dpl_src"

# clean up deployment files
rm -r $dpl_src
