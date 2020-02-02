#!/bin/bash



set -e
cd "$(dirname "$0")"

source ./utils.sh
check_docker_connectivity

BACKENDS=(
    bolt
    mongodb
    swift
    weedfs
)

if [[ -n "$1" ]]; then
    BACKENDS=( "$1" )
fi

if [[ -n "$2" ]]; then
    TEST=$2
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

    if [[ -z "$TEST" ]]; then
        GORACE="halt_on_error=1" go test -count=1 -v -race ../plik/...
    else
        ( cd ../plik && GORACE="halt_on_error=1" go test -count=1 -v -race -run $TEST )
    fi

    #../client/test.sh
    #( cd .. && echo "$PWD" && make docker-make-test )

    $BACKEND/run.sh stop
done