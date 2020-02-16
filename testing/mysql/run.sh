#!/bin/bash

set -e
cd "$(dirname "$0")"

BACKEND="mysql"
CMD=$1
TEST=$2

source ../utils.sh
check_docker_connectivity

DOCKER_IMAGE="mariadb:latest"
DOCKER_NAME="plik.mysql"
DOCKER_PORT=2601

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
        docker run -d -p "$DOCKER_PORT:3306" \
            -e MYSQL_ROOT_PASSWORD="password" \
            -e MYSQL_DATABASE="plik" \
            -e MYSQL_USER="plik" \
            -e MYSQL_PASSWORD="password" \
            --name "$DOCKER_NAME" "$DOCKER_IMAGE"

        echo "waiting for mysqld to start ..."
        sleep 30
        if ! status ; then
            echo "IMAGE IS NOT RUNNING"
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
    #stop
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