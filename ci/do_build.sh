#!/bin/bash -ex

pushd "$(dirname "$0")"/.. > /dev/null
root=$(pwd -P)
popd > /dev/null
export GOPATH=$root/gogo

#----------------------------------------------------------------------

mkdir -p "$GOPATH" "$GOPATH"/bin "$GOPATH"/src "$GOPATH"/pkg

PATH=$PATH:"$GOPATH"/bin

go version

# install metalinter
# go get -u github.com/alecthomas/gometalinter
# gometalinter --install

# build ourself, and go there
go get github.com/venicegeo/dg-pz-workflow
cd $GOPATH/src/github.com/venicegeo/dg-pz-workflow

go test -v github.com/venicegeo/dg-pz-workflow/workflow
