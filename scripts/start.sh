#!/bin/bash
set -eoux pipefail
PID="${HOME}/ni-storage.pid"

./bin/ni-storage & echo $! > "${PID}"