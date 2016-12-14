#!/bin/bash

set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(cd $__dir && cd ../../ && pwd)"
__proj_name="$(basename $__proj_dir)"

export PLUGIN_SRC="${__proj_dir}"
export INFLUXDB_VERSION=${INFLUXDB_VERSION:-"1.0"}

# source the common bash script
. "${__proj_dir}/scripts/common.sh"

# verifies dependencies and starts influxdb
. "${__proj_dir}/examples/tasks/.setup.sh"

# downloads plugins, starts snap, load plugins and start a task
cd "${__proj_dir}/examples/tasks" && docker-compose exec main bash -c "PLUGIN_PATH=/etc/snap/plugins /${__proj_name}/examples/tasks/meminfo-influxdb.sh && printf \"\n\nhint: type 'snaptel task list'\ntype 'exit' when your done\n\n\" && bash"