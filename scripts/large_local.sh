#!/bin/bash

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

export INFLUXDB_VERSION="${INFLUXDB_VERSION:-"v0.10.1"}"
export GOLANGVER="${GOLANGVER:-"1.6.2"}"
export TEST_TYPE="${TEST_TYPE:-"large"}"

docker_folder="${__proj_dir}/scripts/docker"
influx_folder="${docker_folder}/${INFLUXDB_VERSION}"

_debug "docker folder: ${docker_folder}"
_debug "influxdb dockerfile: ${influx_folder}/Dockerfile"

[[ -d ${influx_folder} ]] || _error "invalid Influxdb version: ${INFLUXDB_VERSION}"

_docker_project () {
  (cd "${docker_folder}/${INFLUXDB_VERSION}" && "$@")
}

_debug "building docker image: ${influx_folder}/Dockerfile"
_docker_project docker build -q -t "intelsdi-x/influxdb:${INFLUXDB_VERSION}" .
_debug "running docker image: intelsdi-x/influxdb:${INFLUXDB_VERSION}"
docker_id=$(docker run -d -e PRE_CREATE_DB="test" -p 8083:8083 -p 8086:8086 --expose 8090 --expose 8099 "intelsdi-x/influxdb:${INFLUXDB_VERSION}")

_go_get github.com/smartystreets/goconvey/convey
_go_get github.com/smartystreets/assertions

export SNAP_INFLUXDB_HOST=127.0.0.1
_go_test
_debug "stopping docker image: ${docker_id}"
docker stop "${docker_id}" > /dev/null
_debug "removing docker image: ${docker_id}"
docker rm "${docker_id}" > /dev/null
