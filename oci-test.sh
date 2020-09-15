#!/bin/bash
set -ex

export FN_TEST_MODE="OCI"

# FILL IN VALUES BELOW. These values should match with values in the following files:
# oci-auth/fn/config.yaml
# oci-auth/fn/contexts/functions-test.yaml
# oci-auth/oci/config
# simpleapp/app.json
# docker/config.json
export FN_API_URL="https://functions.us-ashburn-1.oci.oraclecloud.com/20181201"
export FN_SUBNET="ocid1.subnet.oc1.iad.aaaaaaaaxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export FN_IMAGE="iad.ocir.io/registry-name/repo-name/func-image-name:0.0.3"
export FN_IMAGE_2="iad.ocir.io/registry-name/repo-name/another-func-image-name:0.0.1"
export FN_REGISTRY="iad.ocir.io/registry-name/repo-name/"

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
go test -timeout 60m  -v  github.com/fnproject/cli/test
