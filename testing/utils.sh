#! /bin/bash

set -e

ROOT=$(realpath ../..)

function check_docker_connectivity {
    if docker version >/dev/null 2>/dev/null ; then
        true
    else
        echo "Cannot connect to docker daemon."
        if [[ $EUID -ne 0 ]]; then
            echo "Maybe you need to run this as root."
        fi
        false
    fi
}

function build_docker_image {
    ( cd "$ROOT" && make build-docker-builder )
}

function run_tests {
    BACKEND="$1"
    TEST="$2"

    if [[ -z "$BACKEND" ]]; then
        echo "missing backend"
        return 1
    fi

    PLIKD_CONFIG="$ROOT/testing/$BACKEND/plikd.cfg"
    export PLIKD_CONFIG

    if [[ -z "$TEST" ]]; then
        ( cd "$ROOT/plik" && GORACE="halt_on_error=1" go test -count=1 -v -race ./... )
    else
        ( cd "$ROOT/plik" && GORACE="halt_on_error=1" go test -count=1 -v -race -run $TEST )
    fi
}