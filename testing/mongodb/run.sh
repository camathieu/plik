#!/bin/bash

set -e
cd "$(dirname "$0")"

source ../utils.sh

BACKEND="mongodb"
CMD=$1
TEST=$2

DOCKER_COMPOSE_FILE="docker-compose.yml"
DOCKER_COMPOSE_TEST_FILE="docker-compose-test-runner.yml"
DOCKER_NAMES=( "plik.mongodb.rs0.1" "plik.mongodb.rs0.2" "plik.mongodb.rs0.3" )

function start {
    check_docker_connectivity
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d

    local READY=0
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

    local OK=0
    for DOCKER_NAME in "${DOCKER_NAMES[@]}"
    do
        docker exec -t "$DOCKER_NAME" sh -c 'mongo < /scripts/create_mongo_users.js' && OK=1
    done

    if [[ "$OK" == "0" ]]; then
        echo -e "\n - Failed to create MongoDB users\n"
        exit 1
    fi
}

function stop {
    check_docker_connectivity
    docker-compose -f "$DOCKER_COMPOSE_FILE" down
}

function status {
    check_docker_connectivity
    for name in "${DOCKER_NAMES[@]}"
    do
        if docker ps -f name="$name" | grep "$name" > /dev/null ; then
            echo "$name is RUNNING"
        else
            echo "$name is NOT RUNNING"
        fi
    done
}

function run_docker {
    check_docker_connectivity

    workdir="/go/src/github.com/root-gg/plik"
    entrypoint="$1"
    shift

    # Generate version.go
    ( cd "$ROOT" && make build-info )

    docker run -it --rm \
      --network "plik-mongodb-test" \
      --workdir="$workdir" \
      --entrypoint="$entrypoint" \
      --volume="$ROOT:$workdir" \
      -p 8080:8080 \
      golang:latest "$@"
}

# Cleaning shutdown hook
function shutdown {
    echo "CLEANING UP $BACKEND !!!"
    # do not try to run stop inside the docker
    check_docker_connectivity >/dev/null 2>&1 && stop
}
trap shutdown EXIT

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
  status)
    status
    ;;
  run_tests)
    run_tests "$BACKEND" "$TEST"
    ;;
  test)
    stop
    start
    run_docker "./testing/mongodb/run.sh" "run_tests" "$TEST"
    ;;
  dev)
    stop
    start
    run_docker "/bin/bash"
    ;;
  *)
	echo "Usage: $0 {start|stop|restart|status|test [test_name]}"
	exit 1
esac

exit 0