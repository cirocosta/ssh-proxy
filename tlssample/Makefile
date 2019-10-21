build:
	go build -v -i

xbuild:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -v -tags netgo -ldflags '-extldflags "-static"' -o tlssample -i .

