#!/bin/bash
set -eoux pipefail
PID="${HOME}/ni-storage.pid"

if [ -f ${PID} ]
then
    kill -SIGINT $(cat ${PID})
    rm ${PID}
fi