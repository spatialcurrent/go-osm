#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p $DIR/../bin
NAME=go-osm

echo "******************"
echo "Formatting $DIR/../cmd/osm"
cd $DIR/../cmd/osm
go fmt
echo "Formatting github.com/spatialcurrent/$NAME/osm"
go fmt github.com/spatialcurrent/$NAME/osm
echo "Done formatting."
echo "******************"
echo "Building plugin for $NAME"
cd $DIR/../bin
go build github.com/spatialcurrent/$NAME/cmd/osm
if [[ "$?" != 0 ]] ; then
    echo "Error building $NAME program"
    exit 1
fi

echo "Executable built at $(realpath $DIR/../bin/osm)"
