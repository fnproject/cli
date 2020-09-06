#!/bin/bash
set -ex

export FN_TEST_MODE="OCI"

# FILL IN VALUES BELOW. These values should match with values in
# oci-auth/fn/config.yaml
# oci-auth/fn/contexts/functions-test.yaml
# oci-auth/oci/config
export FN_API_URL="https://functions.us-ashburn-1.oraclecloud.com/20181201"
export FN_SUBNET="ocid1.subnet.oc1.iad.aaaaaaaaeclqhptkrlfgmobby2s4g7wz5mj3qvixdrs2twy6gof26qoazbsq"
export FN_IMAGE="iad.ocir.io/odx-jafar/hapai-functions/helloworld-func:0.0.3"
export FN_IMAGE_2="iad.ocir.io/odx-jafar/hapai-functions/simplefunc:0.0.13"
export FN_REGISTRY="iad.ocir.io/odx-jafar/hapai-functions/"

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
