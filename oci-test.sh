#!/bin/bash
set -ex

export FN_TEST_MODE="OCI"

# FILL IN VALUES BELOW. These values should match with values in the following files:
# oci-auth/fn/config.yaml
# oci-auth/fn/contexts/functions-test.yaml
# oci-auth/oci/config
# simpleapp/app.json
export FN_API_URL="https://functions.us-ashburn-1.oraclecloud.com/20181201"
export FN_SUBNET="Fill me in"
export FN_IMAGE="Fill me in"
export FN_IMAGE_2="Fill me in"
export FN_REGISTRY="Fill me in"

function cleanup {
	if [ -d "$WORK_DIR" ]; then
		rm -rf $WORK_DIR
	fi
}
trap cleanup EXIT

CUR_DIR=$(pwd)

export fn="${CUR_DIR}/fn"


#on CI these can take a while
go test -v $(go list ./... |  grep -v "^github.com/fnproject/cli/test$")

# run the CLI ign tests
go test -timeout 40m  -v  github.com/fnproject/cli/test
