#!/bin/bash
#
# Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

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


# start fn
CONTAINER_ID=$($fn start -d | tail -1)

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
go test -timeout 20m  -v  github.com/fnproject/cli/test
