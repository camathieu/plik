#!/bin/bash

###
# The MIT License (MIT)
#
# Copyright (c) <2015>
# - Mathieu Bodjikian <mathieu@bodjikian.fr>
# - Charles-Antoine Mathieu <skatkatt@root.gg>
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
###

set -e
cd "$(dirname "$0")"

source ../utils.sh
check_docker_connectivity

DOCKER_IMAGE="library/mysql:latest"
DOCKER_NAME="plik.mysql"
DOCKER_PORT=2602

MYSQL_ROOT_PASSWORD="password"
MYSQL_USER="plik"
MYSQL_PASSWORD="password"
MYSQL_DATABASE="plik"

function start {
    if status ; then
        echo "ALREADY RUNNING"
        exit 0
    else
        echo -e "\n - Pulling $DOCKER_IMAGE\n"
        docker pull "$DOCKER_IMAGE"
        if docker ps -a -f name="$DOCKER_NAME" | grep "$DOCKER_NAME" > /dev/null ; then
            docker rm -f "$DOCKER_NAME"
        fi

        echo -e "\n - Starting $DOCKER_NAME\n"
        docker run -d -p "$DOCKER_PORT:27017" --name "$DOCKER_NAME" -v "$(pwd):/scripts" \
            -e MYSQL_ROOT_PASSWORD="$MYSQL_ROOT_PASSWORD" \
            -e MYSQL_USER="$MYSQL_USER" \
            -e MYSQL_PASSWORD="$MYSQL_PASSWORD" \
            -e MYSQL_DATABASE="$MYSQL_DATABASE" \
            "$DOCKER_IMAGE"

        echo -e "\n - Waiting for startup ...\n"


        MYSQL_CMD="mysql -u $MYSQL_USER -p$MYSQL_PASSWORD --database $MYSQL_DATABASE"
        for i in $(seq 0 60)
        do
            echo "Waiting for everything to start"
            sleep 1

            DOCKER_ID=$(docker ps -q -f "name=$DOCKER_NAME")
            if [ -z "$DOCKER_ID" ]; then
                echo "Unable to get CONTAINER ID for $DOCKER_NAME"
                exit 1
            fi

            READY="0"
            if docker exec -t "$DOCKER_NAME" sh -c "$MYSQL_CMD -e \"SELECT USER();\"" >/dev/null 2>/dev/null ; then
                READY="1"
                break
            fi
        done

        if [ "$READY" == "1" ]; then
            echo -e "\n - Initializing MySQL\n"
            docker exec -t "$DOCKER_NAME" sh -c "$MYSQL_CMD < /scripts/create.sql"
        else
            echo -e "\n - Unable to connect to MySQL\n"
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

case "$1" in
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