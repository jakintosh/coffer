echo "Installing..."

NAME=${1:?"Service name required."} || exit 1
DOMAIN=${2:?"Domain name required."} || exit 1
DEPLOY_DIR=${3:?"Deployment directory required."} || exit 1

# check if was running for later, stop service
sudo systemctl is-active --quiet $NAME@$DOMAIN
IS_RUNNING=$?

# if service was running, stop it
if [ $IS_RUNNING -eq 0 ]; then
  sudo systemctl stop $NAME@$DOMAIN
fi

sudo mkdir -p /etc/$NAME/$DOMAIN
sudo mkdir -p /var/lib/$NAME/$DOMAIN # for database

sudo cp    $DEPLOY_DIR/usr/local/bin/$NAME  /usr/local/bin/
sudo cp -r $DEPLOY_DIR/etc/systemd/system/. /etc/systemd/system/
sudo cp -r $DEPLOY_DIR/etc/$NAME/.          /etc/$NAME/$DOMAIN

sudo systemctl daemon-reload

# if service was running, start it again
if [ $IS_RUNNING -eq 0 ]; then
  sudo systemctl start $NAME@$DOMAIN.service
fi
