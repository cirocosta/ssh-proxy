build:
	go build -v -o ssh-proxy -i .

image:
	DOCKER_BUILDKIT=1 \
		docker build \
			--target release \
			--tag cirocosta/ssh-proxy .

test:
	DOCKER_BUILDKIT=1 \
		docker build \
			--target test \
			--tag cirocosta/ssh-proxy-test .

fmt:
	go fmt ./...

keys:
	mkdir -p ./keys
	ssh-keygen -t rsa -f ./keys/id_rsa

clean:
	rm -rf ./keys
