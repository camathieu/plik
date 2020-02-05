#!/bin/bash

set -e
cd "$(dirname "$0")"

BACKEND="bolt"
CMD=$1
TEST=$2

source ../utils.sh

case "$CMD" in
  test)
    run_tests "$BACKEND" "$TEST"
    ;;
  *)
	echo "Usage: $0 test [test_name]"
	exit 1
esac

exit 0