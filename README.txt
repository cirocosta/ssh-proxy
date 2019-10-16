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

		ssh-proxy server \
			--private-key=./id_rsa \ 
			--addr=0.0.0.0:2222 \   	# addr to listen for ssh conns
			--port=1234			# port to use to accept conns to forward


	3. on `machine2`, create a "client"

		ssh-proxy client \
			--addr=machine1:2222 \	# addr of the ssh server
			--port=8000		# port to forward conns to
			

		python -m "SimpleHTTPServer"	# start a basic http server here


	4. get requests from the server down to the client

		curl machine1:1234



SHOULD I USE THIS?

	probably not.



LICENSE

	Apache V2 (see ./LICENSE).

	see Concourse's license:
	- https://github.com/concourse/concourse/blob/2cc847226ee5210d30045b49bac7f4c2350990c8/LICENSE.md



