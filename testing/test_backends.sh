#!/bin/bash



set -e
cd "$(dirname "$0")"

source ./utils.sh
check_docker_connectivity

BACKENDS=(
    mongodb
    swift
    weedfs
)

# Cleaning shutdown hook
function shutdown {
    echo "CLEANING UP !!!"
    for BACKEND in "${BACKENDS[@]}"
    do
        echo "CLEANING $BACKEND"
        $BACKEND/run.sh stop
    done
}
trap shutdown EXIT

for BACKEND in "${BACKENDS[@]}"
do
    echo -e "\n - Tesing $BACKEND :\n"

    $BACKEND/run.sh stop
    $BACKEND/run.sh start

    export PLIKD_CONFIG=$(realpath $BACKEND/plikd.cfg)

    GORACE="halt_on_error=1" go test -v -count=1 -race ../plik/...
    #../client/test.sh

    $BACKEND/run.sh stop
done