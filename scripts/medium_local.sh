#!/bin/bash

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

export INFLUXDB_VERSION="${INFLUXDB_VERSION:-"0.13"}"
export TEST_TYPE="${TEST_TYPE:-"medium"}"

docker_folder="${__proj_dir}/config"

_docker_project () {
  (cd "${docker_folder}" && "$@")
}

docker_id=$(docker run -d -e PRE_CREATE_DB="test" -p 8083:8083 -p 8086:8086 "influxdb:${INFLUXDB_VERSION}-alpine")

_debug "creating database"
export SNAP_INFLUXDB_HOST=$(echo $DOCKER_HOST | grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])')
sleep 3 && curl -i -XPOST http://${SNAP_INFLUXDB_HOST}:8086/query --data-urlencode "q=CREATE DATABASE test"

_go_get github.com/smartystreets/goconvey/convey
_go_get github.com/smartystreets/assertions

set +e  # don't bail out of the script without stopping/removing the docker container
_go_test
_debug "stopping docker image: ${docker_id}"
docker stop "${docker_id}" > /dev/null
_debug "removing docker image: ${docker_id}"
docker rm "${docker_id}" > /dev/null
