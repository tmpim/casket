package caskettls

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/caddyserver/certmagic"
	"math/big"
	"net"
	"strings"
	"time"
)

// newSelfSignedCertificate returns a new self-signed certificate.
func newSelfSignedCertificate(ssconfig selfSignedConfig) (tls.Certificate, error) {
	// start by generating private key
	var privKey interface{}
	var err error

	keyGenerator := certmagic.StandardKeyGenerator{KeyType: ssconfig.KeyType}
	privKey, err = keyGenerator.GenerateKey()
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %v", err)
	}

	// create certificate structure with proper values
	notBefore := time.Now()
	notAfter := ssconfig.Expire
	if notAfter.IsZero() || notAfter.Before(notBefore) {
		notAfter = notBefore.Add(24 * time.Hour * 7)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %v", err)
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{Organization: []string{"Casket Self-Signed"}},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	if len(ssconfig.SAN) == 0 {
		ssconfig.SAN = []string{""}
	}
	for _, san := range ssconfig.SAN {
		if ip := net.ParseIP(san); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, strings.ToLower(san))
		}
	}

	// generate the associated public key
	publicKey := func(privKey interface{}) interface{} {
		switch k := privKey.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey
		case *ecdsa.PrivateKey:
			return &k.PublicKey
		default:
			return fmt.Errorf("unknown key type")
		}
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, publicKey(privKey), privKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("could not create certificate: %v", err)
	}

	chain := [][]byte{derBytes}

	return tls.Certificate{
		Certificate: chain,
		PrivateKey:  privKey,
		Leaf:        cert,
	}, nil
}

// selfSignedConfig configures a self-signed certificate.
type selfSignedConfig struct {
	SAN     []string
	KeyType certmagic.KeyType
	Expire  time.Time
}
