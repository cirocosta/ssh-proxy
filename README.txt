WHAT

	this is about testing out the performance differences between SSH and
	Wireguard when it comes to getting traffic through it in terms of:

	- cost (utilization)
	- end-user performance (throughput)


WHY

	Currently, `concourse` gets its workers registered against the cluster
	through SSH, regardless of the topology of the cluster.


			 (ssh)
		worker ----------.
				tsa
		worker ----------*
			 (ssh)
	

	While it's *very* easy to just leverage SSH (being implemented by
	Concourse on both the client and server side) to have workers
	registering from anywhere, there might be a compeling reason to allow
	the case for secure communication without the use of a userspace
	protocol.



THE TOOLS

	SSH
		
		what's that?
		how does it work?
		how Concourse uses it?	


	WIREGUARD

		what's this thing?
		how does it work?
		when did this get possible?
		what are the next steps in its development?
		

METODOLOGY

	> how will we know that one is better than the other?
	> how can we be confident that one is better than the other?

	SCENARIOS

		
		SSH w/ remote port-forwarding


			DIRECT COMM


			s1				s2

			transmitter -> 1234		receiver	:2222
			|						 ^
			|						 |
			ssh server --------------------------------- ssh client
			port-forward 1234:2222


			PROXIED COMM


		
			s1		s2		s3

			transmitter     proxy           receiver




		WIREGUARD


			s1				s2

			transmitter ------------------> receiver	:2222
					wireguard tunnel


	BASELINE

		1. transmit certain payloads with direct communication (no SSH,
		   etc).

		2. perform th

		
	

	in both cases, run the experiments for several payloads, measuring:
	1. cpu utilization during the period (server and client)
	2. egress & ingress in the wire
	3. ram utilized

