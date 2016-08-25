#!/bin/bash

set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# source the common bash script 
. "${__dir}/../../scripts/common.sh"

# ensure PLUGIN_PATH and BIN_PATH are set
TMPDIR=${TMPDIR:-"/tmp"}
PLUGIN_PATH=${PLUGIN_PATH:-"${TMPDIR}/snap/plugins"}
BIN_PATH=${BIN_PATH:-"${TMPDIR}/snap/bin"}
mkdir -p $PLUGIN_PATH
mkdir -p $BIN_PATH

_notice "Get latest snapd, snapctl and plugins"

_debug "BIN_PATH=$BIN_PATH"
_debug "PLUGIN_PATH=$PLUGIN_PATH"

(cd $BIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snapd && chmod 755 snapd)
(cd $BIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snapctl && chmod 755 snapctl)
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snap-plugin-publisher-mock-file && chmod 755 snap-plugin-publisher-mock-file)
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snap-plugin-collector-mock2 && chmod 755 snap-plugin-collector-mock2)
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/plugin/build/latest/snap-plugin-publisher-influxdb && chmod 755 snap-plugin-publisher-influxdb)

_info "creating database"
curl -i -XPOST http://influxdb:8086/query --data-urlencode "q=CREATE DATABASE test"

_info "starting snapd"
$BIN_PATH/snapd -t 0 -l 1 -a $PLUGIN_PATH --log-path /tmp &
snapd_pid=$!

# sleep for a few seconds while we wait for the plugins to load
sleep 5  

_info "creating and starting a task"
$BIN_PATH/snapctl task create -t "${__dir}/task-mock-influxdb.yml" 