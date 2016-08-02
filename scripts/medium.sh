#!/bin/bash

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"
__proj_name="$(basename $__proj_dir)"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

# NOTE: these variables control the docker-compose image.
export INFLUXDB_VERSION="${INFLUXDB_VERSION:-"v0.10.1"}"
export PLUGIN_SRC="${__proj_dir}"

export GOLANGVER="${GOLANGVER:-"1.6.2"}"

TEST_TYPE="${TEST_TYPE:-"medium"}"

docker_folder="${__proj_dir}/scripts/docker/${TEST_TYPE}"
influx_folder="${docker_folder}/../influxdb/${INFLUXDB_VERSION}"

_docker_org_path="\$GOPATH/src/github.com/intelsdi-x"
_docker_proj_path="${_docker_org_path}/${__proj_name}"

_debug "docker folder: ${docker_folder}"
_debug "influxdb dockerfile: ${influx_folder}/Dockerfile"

[[ -d ${influx_folder} ]] || _error "invalid Influxdb version: ${INFLUXDB_VERSION}"

_docker_project () {
  cd "${docker_folder}" && "$@"
}

_debug "building docker compose images"
_docker_project docker-compose build
_debug "running docker compose images"
_docker_project docker-compose up -d
_debug "running test: ${TEST_TYPE}"
_docker_project docker-compose exec -T golang gvm-bash.sh "gvm use $GOLANGVER; export INFLUXDB_VERSION=$INFLUXDB_VERSION; mkdir -p ${_docker_org_path}; cp -Rf /${__proj_name} ${_docker_org_path}; (cd ${_docker_proj_path} && ./scripts/medium_jenkins.sh)"
_debug "stopping docker compose images"
_docker_project docker-compose stop
_docker_project docker-compose rm -f
