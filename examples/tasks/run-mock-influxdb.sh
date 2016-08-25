#!/bin/bash

set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(cd $__dir && cd ../../ && pwd)"
__proj_name="$(basename $__proj_dir)"

export PLUGIN_SRC="${__proj_dir}"

# source the common bash script 
. "${__dir}/../../scripts/common.sh"

# verifies dependencies and starts influxdb
. ${__dir}/.setup.sh

# downloads plugins, starts snap, load plugins and start a task
docker-compose exec main bash -c "PLUGIN_PATH=/etc/snap/plugins BIN_PATH=/usr/local/bin /${__proj_name}/examples/tasks/mock-influxdb.sh && printf \"\n\nhint: type 'snapctl task list'\ntype 'exit' when your done\n\n\" && bash && killall snapd"