ssh-proxy
	

	an 80% copy of Concourse's SSH proxying.



WHY

	Proxying traffic through SSH is definitely not free, and it turns out
	that Concourse does that for streaming bits from one worker to another.

	This was created to isolate the proxying part so that we could gather
	some insights into how costly that is.

	At the same time, this thing is useful on its own for, for instance,
	getting a HTTP server that sits in your desktop (under a private IP
	address) exposed to the "wild interwebzz" by having a server (that sits
	on a public IP) forwarding requests down to your desktop.

		- yeah yeah, maybe not all that useful as I didn't implement any
		  of the key checking for authn hehe



USAGE

	1. create a private key

		ssh-keygen -t rsa -f ./id_rsa


	2. on `machine1`, create a server

		# listening on 2222 for SSH connections
		#
		ssh-proxy server \
			--private-key=./id_rsa \ 
			--addr=0.0.0.0:2222


	3. on `machine2`, create a "client"

		# connects to machine1:2222, sending a request for it
		# to take connections to its 1234 (i.e., `machine1:1234`),
		# forward them to the client, which is then responsible for
		# pushing down to a server listening on 8000.
		#
		ssh-proxy client \
			--addr=machine1:2222 \	# addr of the ssh server
			--local-port=8000 \
			--remote-port=1234
			

		python -m "SimpleHTTPServer"	# start a basic http server here
						# that listens on 8000


	4. get requests from the server down to the client

		curl machine1:1234


			curl --> server:1234 ---> client ---> application:8000



SHOULD I USE THIS?

	probably not.

	this is just for an experiment.



LICENSE

	Apache V2 (see ./LICENSE).

	see Concourse's license:
	- https://github.com/concourse/concourse/blob/2cc847226ee5210d30045b49bac7f4c2350990c8/LICENSE.md



