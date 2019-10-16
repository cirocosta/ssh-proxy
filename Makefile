build:
	go build -v -o ssh-proxy -i .

fmt:
	go fmt ./...

keys:
	mkdir -p ./keys
	ssh-keygen -t rsa -f ./keys/id_rsa

clean:
	rm -rf ./keys
