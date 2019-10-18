#!/bin/bash

# integration - a quick integration test to verify that we got it right.
#

set -o errexit
set -o nounset
set -o xtrace

pids=""

main () {
	run_http_server
	run_ssh_server
	run_ssh_client

	validate
	sleep 1s
}

validate () {
	curl --fail-early --show-error http://localhost:1234
}

run_http_server () {
	dumbserver -addr=:8000 &
	echo "$!"
	sleep 1s
}

run_ssh_client () {
	ssh-proxy client \
		--addr="localhost:2222" \
		--local-port="8000" \
		--remote-port="1234" &
	echo "$!"
	sleep 2s
}

run_ssh_server () {
	ssh-proxy server --private-key=./keys/id_rsa &
	echo "$!"
	sleep 2s
}

main "$@"
