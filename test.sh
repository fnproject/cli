#!/bin/bash
set -ex

function cleanup {
	if [ ! -z "$CONTAINER_ID" ]; then
		docker logs ${CONTAINER_ID}
		docker kill ${CONTAINER_ID}
	fi
	if [ -d "$WORK_DIR" ]; then
		rm -rf $WORK_DIR
	fi
}
trap cleanup EXIT

CUR_DIR=$(pwd)

export fn="${CUR_DIR}/fn"
export FN_REGISTRY=$DOCKER_USER
if [[ -z "$FN_REGISTRY" ]]; then
  export FN_REGISTRY=default_docker_user_does_not_push
fi
$fn --version

export FN_API_URL="http://localhost:8080"

go test $(go list ./... | grep -v /vendor/ | grep -v /tests)

# Our test directory
WORK_DIR=$(mktemp -d)
cd $WORK_DIR

# HACK: if we don't pre touch these, fn server will create these as 'root'
mkdir data
touch data/fn.db data/fn.mq

# start fn
CONTAINER_ID=$($fn start -d)
sleep 7
docker inspect -f {{.State.Running}} $CONTAINER_ID | grep '^true$'

# This tests all the quickstart commands on the cli on a live server
cd $WORK_DIR
funcname="fn-test-go"
mkdir $funcname
cd $funcname
$fn init --runtime go
$fn run
$fn test
$fn apps l
$fn apps create myapp
$fn apps l
$fn -v deploy --local --app myapp
$fn call myapp $funcname

# Test ruby func
cd $WORK_DIR
funcname="rubyfunc"
mkdir $funcname
cd $funcname
$fn init --runtime ruby
$fn run
$fn test

# Test 'docker' runtime deploy
cd $WORK_DIR
funcname="dockerfunc"
mkdir $funcname 
cp ${CUR_DIR}/test/funcfile-docker-rt-tests/testfiles/Dockerfile $funcname/
cp ${CUR_DIR}/test/funcfile-docker-rt-tests/testfiles/func.go $funcname/
cd $funcname
$fn init --name $funcname
$fn apps create myapp1
$fn apps l
$fn -v deploy --local --app myapp1
$fn call myapp1 /$funcname
# todo: would be nice to have a flag to output parseable formats in cli, eg: `fn deploy --output json` would return json with version and other info 
$fn routes create myapp1 /another --image $FN_REGISTRY/$funcname:0.0.2
$fn call myapp1 /another
