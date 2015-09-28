#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2015 Intel Coporation
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

#!/bin/bash -e

GITVERSION=`git describe --always`
SOURCEDIR=$1
BUILDDIR=$SOURCEDIR/build
PLUGIN=`echo $SOURCEDIR | grep -oh "pulse-.*"`
ROOTFS=$BUILDDIR/rootfs
BUILDCMD='go build -a -ldflags "-w"'

echo
echo "****  Pulse Plugin Build  ****"
echo

# Disable CGO for builds
export CGO_ENABLED=0

# Clean build bin dir
rm -rf $ROOTFS/*

# Make dir
mkdir -p $ROOTFS

# Build plugin
echo "Source Dir = $SOURCEDIR"
echo "Building Pulse Plugin: $PLUGIN"
$BUILDCMD -o $ROOTFS/$PLUGIN

