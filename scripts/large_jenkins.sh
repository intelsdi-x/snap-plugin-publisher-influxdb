#!/bin/bash

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"
. "${__dir}/deps.sh"

_go_get github.com/smartystreets/goconvey/convey
_go_get github.com/smartystreets/assertions

export TEST_TYPE="${TEST_TYPE:-"large"}"
export SNAP_INFLUXDB_HOST="influxdb"
_go_test
