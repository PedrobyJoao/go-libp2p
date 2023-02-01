package libp2pwebrtc

import (
	"crypto"
	"crypto/x509"
	"errors"
)

// Fingerprint is forked from pion to avoid bytes to string alloc,
// and to avoid the entire hex interpersing when we do not need it anyway

var (
	errHashUnavailable = errors.New("fingerprint: hash algorithm is not linked into the binary")
)

// Fingerprint creates a fingerprint for a certificate using the specified hash algorithm
func Fingerprint(cert *x509.Certificate, algo crypto.Hash) ([]byte, error) {
	if !algo.Available() {
		return nil, errHashUnavailable
	}
	h := algo.New()
	for i := 0; i < len(cert.Raw); {
		n, _ := h.Write(cert.Raw[i:])
		// Hash.Writer is specified to be never returning an error.
		// https://golang.org/pkg/hash/#Hash
		i += n
	}
	return h.Sum(nil), nil
}
