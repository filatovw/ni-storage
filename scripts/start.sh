#!/bin/bash
set -eoux pipefail
PID="${HOME}/ni-storage.pid"

if [ ! -f "${PID}" ]; then
    ./bin/ni-storage & echo $! > "${PID}"
fi