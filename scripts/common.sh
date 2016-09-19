#!/bin/bash
# File managed by pluginsync

# http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
# Copyright 2016 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set -u
set -o pipefail

LOG_LEVEL="${LOG_LEVEL:-6}"
NO_COLOR="${NO_COLOR:-}"

trap_exitcode() {
  exit $?
}

trap trap_exitcode SIGINT

_fmt () {
  local color_debug="\x1b[35m"
  local color_info="\x1b[32m"
  local color_notice="\x1b[34m"
  local color_warning="\x1b[33m"
  local color_error="\x1b[31m"
  local colorvar=color_$1

  local color="${!colorvar:-$color_error}"
  local color_reset="\x1b[0m"
  if [ "${NO_COLOR}" = "true" ] || [[ "${TERM:-}" != "xterm"* ]] || [ -t 1 ]; then
    # Don't use colors on pipes or non-recognized terminals
    color=""; color_reset=""
  fi
  echo -e "$(date -u +"%Y-%m-%d %H:%M:%S UTC") ${color}$(printf "[%9s]" "${1}")${color_reset}";
}

_debug ()   { [ "${LOG_LEVEL}" -ge 7 ] && echo "$(_fmt debug) ${*}" 1>&2 || true; }
_info ()    { [ "${LOG_LEVEL}" -ge 6 ] && echo "$(_fmt info) ${*}" 1>&2 || true; }
_notice ()  { [ "${LOG_LEVEL}" -ge 5 ] && echo "$(_fmt notice) ${*}" 1>&2 || true; }
_warning () { [ "${LOG_LEVEL}" -ge 4 ] && echo "$(_fmt warning) ${*}" 1>&2 || true; }
_error ()   { [ "${LOG_LEVEL}" -ge 3 ] && echo "$(_fmt error) ${*}" 1>&2 || true; exit 1; }

_test_dirs() {
  local test_dirs=$(find . -type f -name '*.go' -not -path "./.*" -not -path "*/_*" -not -path "./Godeps/*" -not -path "./vendor/*" -print0 | xargs -0 -n1 dirname| sort -u)
  _debug "go code directories ${test_dirs}"
  echo "${test_dirs}"
}

_go_get() {
  local _url=$1
  local _util

  _util=$(basename "${_url}")

  type -p "${_util}" > /dev/null || go get "${_url}" && _debug "go get ${_util} ${_url}"
}

_goimports() {
  _go_get golang.org/x/tools/cmd/goimports
  test -z "$(goimports -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") | tee /dev/stderr)"
}

_golint() {
  _go_get github.com/golang/lint/golint
  golint ./...
}

_go_vet() {
  go vet $(_test_dirs)
}

_go_race() {
  go test -race ./...
}

_go_test() {
  _info "running test type: ${TEST_TYPE}"
  # Standard go tooling behavior is to ignore dirs with leading underscors
  for dir in $(_test_dirs);
  do
    if [[ -z ${go_cover+x} ]]; then
      _debug "running go test with cover in ${dir}"
      go test -v --tags="${TEST_TYPE}" -covermode=count -coverprofile="${dir}/profile.tmp" "${dir}"
      if [ -f "${dir}/profile.tmp" ]; then
        tail -n +2 "${dir}/profile.tmp" >> profile.cov
        rm "${dir}/profile.tmp"
      fi
    else
      _debug "running go test without cover in ${dir}"
      go test -v --tags="${TEST_TYPE}" "${dir}"
    fi
  done
}

_go_cover() {
  go tool cover -func profile.cov
}
