#!/bin/bash

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"
__proj_name="$(basename $__proj_dir)"

. "${__dir}/common.sh"

# NOTE: these variables control the docker-compose image.
export INFLUXDB_VERSION="${INFLUXDB_VERSION:-"v0.10.1"}"
export PLUGIN_SRC="${__proj_dir}"
export LOG_LEVEL="${LOG_LEVEL:-"7"}"
export PROJECT_NAME="${__proj_name}"

TEST_TYPE="${TEST_TYPE:-"large"}"

docker_folder="${__proj_dir}/scripts/docker/${TEST_TYPE}"
influx_folder="${docker_folder}/../influxdb/${INFLUXDB_VERSION}"

[[ -d ${influx_folder} ]] || _error "invalid Influxdb version: ${INFLUXDB_VERSION}"

_docker_project () {
  cd "${docker_folder}" && "$@"
}

_debug "building docker compose images"
_docker_project docker-compose build
_debug "running docker compose images"
_docker_project docker-compose up -d
_debug "running test: ${TEST_TYPE}"
cd "${docker_folder}"

set +e
docker-compose exec main bash -c "export INFLUXDB_VERSION=$INFLUXDB_VERSION; export LOG_LEVEL=$LOG_LEVEL; /${__proj_name}/scripts/large_tests.sh" 
test_res=$?
set -e
_debug "exit code $test_res"
_debug "stopping docker compose images"
_docker_project docker-compose stop
_debug "removing docker compose images"
_docker_project docker-compose rm -f
exit $test_res