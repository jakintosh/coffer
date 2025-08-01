echo "Packaging..."

NAME=${1:?"Service name required."} || exit 1
DEPLOY_DIR=${2:?"Deployment directory required."} || exit 1

mkdir -p $DEPLOY_DIR/usr/local/bin
mkdir -p $DEPLOY_DIR/etc/systemd/system
mkdir -p $DEPLOY_DIR/etc/$NAME

cp    ./scripts/install.sh $DEPLOY_DIR/
cp    ./bin/coffer         $DEPLOY_DIR/usr/local/bin/
cp -r ./init/.             $DEPLOY_DIR/etc/systemd/system/
cp -r ./secrets/.          $DEPLOY_DIR/etc/$NAME/
