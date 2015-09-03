#!/bin/bash -e
# This script runs the correct godep sequences for pulse and built-in plugins
# This will rebase back to the committed version. It should be run from pulse/.
ctrl_c()
{
  exit $?
} 
trap ctrl_c SIGINT

# First load pulse deps
echo "Checking deps for plugin"
godep restore
