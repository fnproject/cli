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


#on CI these can take a while
go test -v $(go list ./... |  grep -v "^github.com/fnproject/cli/test$")

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
while [ true ]; do

	set +e
	curl -sS --max-time 1 ${FN_API_URL} > /dev/null
	RESULT=$?
	set -e

	if [ ${RESULT} -eq 0 ]; then
		break
	else
		echo "Going to try ${FN_API_URL} one more time..."
	fi

	TRIES=$((${TRIES}-1))
	if [ ${TRIES} -le 0 ]; then
		echo "Max retries reached, cannot connect to ${FN_API_URL}"
		exit 1
	fi
	sleep 1
done

# be safe, check the fn container too.
docker inspect -f {{.State.Running}} $CONTAINER_ID | grep '^true$'

# run the CLI ign tests
go test -timeout 20m  -v  github.com/fnproject/cli/test
