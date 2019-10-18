#!/bin/bash

set -o errexit
set -o nounset

readonly NE_VERSION=0.18.1
readonly NE_URL="https://github.com/prometheus/node_exporter/releases/download/v$NE_VERSION/node_exporter-$NE_VERSION.linux-amd64.tar.gz"

main () {
	install_node_exporter
}

install_node_exporter () {
	curl -SOL $NE_URL 
	tar xvzf ./node_exporter-$NE_VERSION.linux-amd64.tar.gz \
		node_exporter-$NE_VERSION.linux-amd64/node_exporter \
		--strip-components=1
	install -m 0755 ./node_exporter /usr/local/bin
}

main "$@"
