#!/bin/bash

set -e
cd "$(dirname "$0")"

source ../utils.sh
check_docker_connectivity

DOCKER_COMPOSE_FILE="docker-compose.yml"
DOCKER_NAMES=( "plik.mongodb.rs0.1" "plik.mongodb.rs0.2" "plik.mongodb.rs0.3" )

function start {
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d

    READY=0
    for i in $(seq 0 60)
    do
        echo "Waiting for everything to start"
        sleep 1

        READY=0
        for DOCKER_NAME in "${DOCKER_NAMES[@]}"
        do
            DOCKER_ID=$(docker ps -q -f "name=$DOCKER_NAME")
            if [ -z "$DOCKER_ID" ]; then
                echo "Unable to get CONTAINER ID for $DOCKER_NAME"
                exit 1
            fi

            if docker exec -t "$DOCKER_NAME" sh -c "mongo --eval \"version();\"" >/dev/null 2>/dev/null ; then
                READY=$((READY +1))
            fi
        done

        if [ "$READY" == "${#DOCKER_NAMES[@]}" ]; then
          break
        fi
    done

    if [ "$READY" != "${#DOCKER_NAMES[@]}" ]; then
        echo -e "\n - Unable to connect to MongoDB\n"
        exit 1
    fi

    echo -e "\n - Initializing MongoDB replica set\n"
    docker exec -t "${DOCKER_NAME[0]}" sh -c 'mongo < /scripts/create_mongo_replica_set.js'

    sleep 1

    echo -e "\n - Initializing MongoDB users\n"
    for DOCKER_NAME in "${DOCKER_NAMES[@]}"
    do
        docker exec -t "$DOCKER_NAME" sh -c 'mongo < /scripts/create_mongo_users.js' && exit 0
    done

    echo -e "\n - Failed to create MongoDB users\n"
    exit 1
}

function stop {
    docker-compose -f "$DOCKER_COMPOSE_FILE" down
}

function status {
    for name in "${DOCKER_NAMES[@]}"
    do
        if docker ps -f name="$name" | grep "$name" > /dev/null ; then
            echo "$name is RUNNING"
        else
            echo "$name is NOT RUNNING"
        fi
    done
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
    status
    ;;
  *)
	echo "Usage: $0 {start|stop|restart|status}"
	exit 1
esac

exit 0