#!/bin/bash  

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

. "${__dir}/common.sh"

return_code=0

_notice "Get latest snapd, snapctl and mock plugins"

mkdir -p /etc/snap/plugins

(cd /usr/local/bin && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snapd && chmod 755 /usr/local/bin/snapd)
(cd /usr/local/bin/ && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snapctl && chmod 755 /usr/local/bin/snapctl)
(cd /etc/snap/plugins && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snap-plugin-publisher-mock-file && chmod 755 /etc/snap/plugins/snap-plugin-publisher-mock-file)
(cd /etc/snap/plugins && curl -sSO http://snap.ci.snap-telemetry.io/snap/master/latest/snap-plugin-collector-mock1 && chmod 755 /etc/snap/plugins/snap-plugin-collector-mock1)
(cd /etc/snap/plugins && curl -sSO http://snap.ci.snap-telemetry.io/plugin/build/latest/snap-plugin-publisher-influxdb && chmod 755 /etc/snap/plugins/snap-plugin-publisher-influxdb)

_info "starting snapd"
/usr/local/bin/snapd -t 0 -l 1 -a /etc/snap/plugins --log-path /tmp &
snapd_pid=$!

_info "waiting 5 seconds for snap to start and plugins to load" && sleep 5

_info "creating and starting a task"
/usr/local/bin/snapctl task create -t "${__proj_dir}/scripts/docker/large/mock-influxdb.yaml"

_debug "Sleeping for 10 seconds so the task can do some work"
sleep 10

echo -n "[task is running] "
task_list=$(snapctl task list | tail -1)
if echo $task_list | grep -q Running; then
    echo "ok"
else 
    echo "not ok"
    return_code=-1
fi

echo -n "[task is hitting] "
if [ $(echo $task_list | awk '{print $4}') -gt 0 ]; then
    echo "ok"
else 
    _debug $task_list
    echo "not ok"
    return_code=-1
fi

echo -n "[task has no errors] "
if [ $(echo $task_list | awk '{print $6}') -eq 0 ]; then
    echo "ok"
else 
    echo "not ok"
    return_code=-1
fi

kill $snapd_pid

exit $return_code
