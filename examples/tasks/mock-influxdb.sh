#!/bin/bash

set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# source the common bash script 
. "${__dir}/../../scripts/common.sh"

# ensure PLUGIN_PATH is set
TMPDIR=${TMPDIR:-"/tmp"}
PLUGIN_PATH=${PLUGIN_PATH:-"${TMPDIR}/snap/plugins"}
mkdir -p $PLUGIN_PATH

_info "Get latest plugins"
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snap-plugin-publisher-mock-file && chmod 755 snap-plugin-publisher-mock-file)
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snap-plugin-collector-mock2 && chmod 755 snap-plugin-collector-mock2)
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/plugin/build/latest/snap-plugin-publisher-influxdb && chmod 755 snap-plugin-publisher-influxdb)

_info "creating database"
curl -i -XPOST http://influxdb:8086/query --data-urlencode "q=CREATE DATABASE test"

_info "loading plugins"
snapctl plugin load "${PLUGIN_PATH}/snap-plugin-publisher-mock-file"
snapctl plugin load "${PLUGIN_PATH}/snap-plugin-collector-mock2"
snapctl plugin load "${PLUGIN_PATH}/snap-plugin-publisher-influxdb" 

_info "creating and starting a task"
snapctl task create -t "${__dir}/task-mock-influxdb.yml" 