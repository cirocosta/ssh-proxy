#!/bin/bash

# integration - a quick integration test to verify that we got it right.
#

set -o errexit
set -o nounset

pids=""

main () {
	pids="$pids $(run_http_server)"
	pids="$pids $(run_ssh_server)"
	pids="$pids $(run_ssh_client)"

	trap teardown EXIT

	validate
}

teardown () {
	for pid in pids; do
		kill $pid
	done
}

validate () {
	curl --fail-early --show-error http://localhost:1234
}

run_http_server () {
	python -m "SimpleHTTPServer" &
	echo "$!"
	sleep 3
}

run_ssh_client () {
	./forwarder connect --addr="localhost:2222" --port="8000" &
	echo "$!"
	sleep 3
}

run_ssh_server () {
	./forwarder serve --port="1234" &
	echo "$!"
	sleep 3
}

main "$@"
