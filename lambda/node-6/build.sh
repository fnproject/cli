set -ex

docker build --build-arg HTTP_PROXY -t fnproject/lambda:node-6 .
