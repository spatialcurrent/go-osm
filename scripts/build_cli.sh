#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p $DIR/../bin

echo "******************"
echo "Formatting $DIR/osm"
cd $DIR/../osm
go fmt
echo "Formatting $DIR/../cmd/osm"
cd $DIR/../cmd/osm
go fmt
echo "Done formatting."
echo "******************"
echo "Building program osm"
cd $DIR/../bin
####################################################
#echo "Building program for darwin"
#GOTAGS= CGO_ENABLED=1 GOOS=${GOOS} GOARCH=amd64 go build --tags "darwin" -o "osm_darwin_amd64" github.com/spatialcurrent/go-osm/cmd/osm
#if [[ "$?" != 0 ]] ; then
#    echo "Error building osm for Darwin"
#    exit 1
#fi
#echo "Executable built at $(realpath $DIR/../bin/osm_darwin_amd64)"
####################################################
echo "Building program for linux"
GOTAGS= CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build --tags "linux" -o "osm_linux_amd64" github.com/spatialcurrent/go-osm/cmd/osm
if [[ "$?" != 0 ]] ; then
    echo "Error building osm for Linux"
    exit 1
fi
echo "Executable built at $(realpath $DIR/../bin/osm_linux_amd64)"
####################################################
echo "Building program for Windows"
GOTAGS= CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o "osm_windows_amd64.exe" github.com/spatialcurrent/go-osm/cmd/osm
if [[ "$?" != 0 ]] ; then
    echo "Error building osm for Windows"
    exit 1
fi
echo "Executable built at $(realpath $DIR/../bin/osm_windows_amd64.exe)"
