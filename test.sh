set -ex

make build
export fn="$(pwd)/fn"
export FN_REGISTRY=$DOCKER_USER
$fn --version

go test $(go list ./... | grep -v /vendor/ | grep -v /tests)

# This tests all the quickstart commands on the cli on a live server
rm -rf tmp
mkdir tmp
cd tmp
funcname="fn-test-go"
$fn init --runtime go --name $funcname
$fn test

someport=50080
docker rm --force functions || true # just in case
docker pull fnproject/functions
docker run --name functions -d -v /var/run/docker.sock:/var/run/docker.sock -p $someport:8080 fnproject/functions
sleep 10
docker logs functions

export API_URL="http://localhost:$someport"
$fn apps l
$fn apps create myapp
$fn apps l
$fn deploy --local myapp
$fn call myapp $funcname

#Test 'docker' runtime deploy
mkdir  docker_runtime_test 
cp ../test/funcfile-docker-rt-tests/testfiles/Dockerfile docker_runtime_test/
cp ../test/funcfile-docker-rt-tests/testfiles/func.go docker_runtime_test/
cd docker_runtime_test
$fn init --name $funcname
$fn apps create myapp1
$fn apps l
$fn deploy --local --app myapp1
$fn routes create myapp1 /$funcname
$fn call myapp1 /$funcname

docker rm --force functions

cd ../..
