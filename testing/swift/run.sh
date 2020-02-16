#!/bin/bash

set -e
cd "$(dirname "$0")"

BACKEND="swift"
CMD=$1
TEST=$2

source ../utils.sh
check_docker_connectivity

# rootgg/swift is a build of https://github.com/ccollicutt/docker-swift-onlyone

DOCKER_IMAGE="rootgg/swift:latest"
DOCKER_NAME="plik.swift"
DOCKER_PORT=2603
#SWIFT_DIRECTORY="/tmp/plik.swift.tmpdir"

function start {
    if status ; then
        echo "ALREADY RUNNING"
    else
        echo -e "\n - Pulling $DOCKER_IMAGE\n"
        docker pull "$DOCKER_IMAGE"
        if docker ps -a -f name="$DOCKER_NAME" | grep "$DOCKER_NAME" > /dev/null ; then
            docker rm -f "$DOCKER_NAME"
        fi

        echo -e "\n - Starting $DOCKER_NAME\n"
        docker run -d -p "$DOCKER_PORT:8080" --name "$DOCKER_NAME" "$DOCKER_IMAGE"

        for i in $(seq 0 30)
        do
            echo "Waiting for everything to start"
            sleep 1

            DOCKER_ID=$(docker ps -q -f "name=$DOCKER_NAME")
            if [ -z "$DOCKER_ID" ]; then
                echo "Unable to get CONTAINER ID for $DOCKER_NAME"
                exit 1
            fi

            READY="0"
            if curl -s --max-time 1 "http://127.0.0.1:$DOCKER_PORT/info" >/dev/null 2>/dev/null ; then
                READY="1"
                break
            fi
        done

        if [ "$READY" == "1" ]; then
            echo -e "\n - Initializing Swift\n"
            ./initialize.sh
        else
            echo -e "\n - Unable to connect to Swift\n"
            exit 1
        fi
    fi
}

function stop {
    if status ; then
        echo -e "\n - Removing $DOCKER_NAME\n"
        docker rm -f "$DOCKER_NAME" >/dev/null
    else
        echo "NOT RUNNING"
        exit 0
    fi
}

function status {
    docker ps -f name="$DOCKER_NAME" | grep "$DOCKER_NAME" > /dev/null
}

case "$CMD" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  restart)
    stop
    start
    ;;
  test)
    start
    run_tests "$BACKEND" "$TEST"
    ;;
  status)
    if status ; then
        docker ps -f name="$DOCKER_NAME"
    else
        echo "NOT RUNNING"
    fi
    ;;
  *)
	echo "Usage: $0 {start|stop|restart|status}"
	exit 1
esac

exit 0