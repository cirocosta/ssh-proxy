# syntax = docker/dockerfile:experimental

FROM golang AS golang


FROM golang AS base

	ENV CGO_ENABLED=0
	RUN apt update && apt install -y git

	ADD . /src
	WORKDIR /src



FROM base AS build

	RUN \
		--mount=type=cache,target=/root/.cache/go-build \
		--mount=type=cache,target=/go/pkg/mod \
			go build -v \
			-tags "netgo" \
			-o /bin/ssh-proxy \
			-ldflags "-X main.version=$(cat ./VERSION) -extldflags \"-static\""


FROM build AS test


	RUN set -x && \
		mkdir -p keys && \
		ssh-keygen -t rsa -f ./keys/id_rsa -q -N ""

	RUN \
		--mount=type=cache,target=/root/.cache/go-build \
		--mount=type=cache,target=/go/pkg/mod \
			go build -v \
			-tags "netgo" \
			-o /bin/dumbserver \
			-ldflags "-extldflags \"-static\"" ./dumbserver
	RUN ./test.sh


FROM ubuntu AS release

	COPY \
		--from=build \
		/bin/ssh-proxy \
		/usr/local/bin/ssh-proxy

	ENTRYPOINT [ "/usr/local/bin/ssh-proxy" ]
