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


go test -v $(go list ./... |  grep -v github.com/fnproject/cli/test)

# Our test directory
OS=$(uname -s)
if [ $OS = "Darwin" ]; then
	WORK_DIR=$(mktemp -d /tmp/temp.XXXXXX)
else
	WORK_DIR=$(mktemp -d)
fi
cd $WORK_DIR

# HACK: if we don't pre touch these, fn server will create these as 'root'
mkdir data
touch data/fn.db data/fn.mq

# start fn
CONTAINER_ID=$($fn start -d)

FN_API_URL="localhost:8080"

TRIES=15
while [ ${TRIES} -gt 0 ]; do

	set +e
	curl -sS --max-time 1 ${FN_API_URL} > /dev/null
	RESULT=$?
	set -e

	if [ ${RESULT} -ne 0 ]; then
		sleep 1
		TRIES=$((${TRIES}-1))
	fi
	break
done
# exhausted all tries?
test ${TRIES} -gt 0

# be safe, check the fn container too.
docker inspect -f {{.State.Running}} $CONTAINER_ID | grep '^true$'

# run the CLI ign tests
go test -v  github.com/fnproject/cli/test
