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

if [[ -n "$1" ]]; then
    BACKENDS=( "$1" )
fi

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
    if [[ ! -d $BACKEND ]];then
        echo -e "\n invalid backend $BACKEND\n"
        exit 1
    fi

    echo -e "\n - Tesing $BACKEND :\n"

    $BACKEND/run.sh stop
    $BACKEND/run.sh start

    PLIKD_CONFIG=$(realpath "$BACKEND/plikd.cfg")
    export PLIKD_CONFIG

    #GORACE="halt_on_error=1" go test -v -count=1 -race ../plik/...
    #../client/test.sh
    #( cd .. && echo "$PWD" && make docker-make-test )

    $BACKEND/run.sh stop
done