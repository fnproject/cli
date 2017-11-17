set -ex

./build.sh

docker push fnproject/lambda:node-6
