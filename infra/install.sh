#!/bin/bash

set -o errexit
set -o nounset

readonly NE_VERSION=0.18.1
readonly NE_URL="https://github.com/prometheus/node_exporter/releases/download/v$NE_VERSION/node_exporter-$NE_VERSION.linux-amd64.tar.gz"

readonly CG_URL="https://github.com/cirocosta/chicken-gun/releases/download/0.0.1-rc1/cg"

main () {
	install_node_exporter
	install_cg
}

install_cg () {
	curl -SOL $CG_URL
	sudo install -m 0755 ./cg /usr/local/bin
}

install_node_exporter () {
	curl -SOL $NE_URL 
	tar xvzf ./node_exporter-$NE_VERSION.linux-amd64.tar.gz \
		node_exporter-$NE_VERSION.linux-amd64/node_exporter \
		--strip-components=1
	sudo install -m 0755 ./node_exporter /usr/local/bin
	sudo cp ./node_exporter.service /etc/systemd/system/node_exporter.service
	sudo systemctl daemon-reload
	sudo service node_exporter start
}

main "$@"
