#!/bin/bash



set -e
cd "$(dirname "$0")"

source ../utils.sh
check_docker_connectivity

DOCKER_IMAGE="library/mongo:latest"
DOCKER_NAME="plik.mongodb"
DOCKER_PORT=2601

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
        docker run -d -p "$DOCKER_PORT:27017" --name "$DOCKER_NAME" -v "$(pwd):/scripts" "$DOCKER_IMAGE" --auth

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
            if docker exec -t "$DOCKER_NAME" sh -c "mongo --eval \"version();\"" >/dev/null 2>/dev/null ; then
                READY="1"
                break
            fi
        done

        if [ "$READY" == "1" ]; then
            echo -e "\n - Initializing MongoDB\n"
            docker exec -t "$DOCKER_NAME" sh -c 'mongo < /scripts/create_mongo_users.js'
        else
            echo -e "\n - Unable to connect to MongoDB\n"
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