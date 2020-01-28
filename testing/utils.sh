#! /bin/bash

set -e

function check_docker_connectivity {
    if ! docker version >/dev/null 2>/dev/null ; then
        echo "Cannot connect to docker daemon."
        if [[ $EUID -ne 0 ]]; then
            echo "Maybe you need to run this as root."
        fi
        exit 1
    fi
}