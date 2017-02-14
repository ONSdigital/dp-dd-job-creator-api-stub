#!/bin/bash -eux

export BINPATH=$(pwd)/bin
export GOPATH=$(pwd)/go

pushd $GOPATH/src/github.com/ONSdigital/dp-dd-job-creator-api-stub
  go build -o $BINPATH/dp-dd-job-creator-api-stub
popd
