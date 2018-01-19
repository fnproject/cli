set -ex

make build
export fn="$(pwd)/fn"
export FN_REGISTRY=$DOCKER_USER
if [[ -z "$FN_REGISTRY" ]]; then
  export FN_REGISTRY=default_docker_user_does_not_push
fi
$fn --version

go test $(go list ./... | grep -v /vendor/ | grep -v /tests)

# This tests all the quickstart commands on the cli on a live server
rm -rf tmp
mkdir tmp
cd tmp
funcname="fn-test-go"
mkdir $funcname
cd $funcname
$fn init --runtime go
$fn run
$fn test

someport=50080
docker rm --force functions || true # just in case
docker pull fnproject/functions
docker run --name functions -d -v /var/run/docker.sock:/var/run/docker.sock -p $someport:8080 fnproject/functions
sleep 10
docker logs functions

export FN_API_URL="http://localhost:$someport"
$fn apps l
$fn apps create myapp
$fn apps l
$fn deploy --local --app myapp
$fn call myapp $funcname
cd ..

# Test ruby func
funcname="rubyfunc"
mkdir $funcname
cd $funcname
$fn init --runtime ruby
$fn run
$fn test
cd ..

# Test 'docker' runtime deploy
funcname="dockerfunc"
mkdir $funcname 
cp ../test/funcfile-docker-rt-tests/testfiles/Dockerfile $funcname/
cp ../test/funcfile-docker-rt-tests/testfiles/func.go $funcname/
cd $funcname
$fn init --name $funcname
$fn apps create myapp1
$fn apps l
$fn deploy --local --app myapp1
$fn call myapp1 /$funcname
# todo: would be nice to have a flag to output parseable formats in cli, eg: `fn deploy --output json` would return json with version and other info 
$fn routes create myapp1 /another --image $DOCKER_USER/$funcname:0.0.2
$fn call myapp1 /another

docker rm --force functions

cd ../..
rm -rf tmp
