package ssh

import (
	"golang.org/x/crypto/ssh"
)

// defaultSSHConfig is the default SSH configuration shared between client and
// server.
//
var defaultSSHConfig = ssh.Config{
	Ciphers: nil,

	MACs: []string{
		"hmac-sha2-256-etm@openssh.com",
		"hmac-sha2-256",
	},

	KeyExchanges: []string{
		"ecdh-sha2-nistp256",
		"ecdh-sha2-nistp384",
		"ecdh-sha2-nistp521",
		"curve25519-sha256@libssh.org",
	},
}
