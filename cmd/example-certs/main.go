package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	if err := realMain(); err != nil {
		log.Fatalln(err)
	}
}

func realMain() error {
	now := time.Now()
	ca, err := createRootCA(now)
	if err != nil {
		return err
	}
	err = ca.createCert("server", now)
	if err != nil {
		return err
	}
	err = ca.createCert("alice", now)
	if err != nil {
		return err
	}
	err = ca.createCert("bob", now)
	if err != nil {
		return err
	}
	return nil
}

type rootCA struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
}

func createRootCA(now time.Time) (*rootCA, error) {
	log.Println("generating root ca")
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2023),
		Subject: pkix.Name{
			CommonName:    "example root ca",
			Country:       []string{"AU"},
			Organization:  []string{"example grpc"},
			Locality:      []string{"Brisbane"},
			Province:      []string{"Queensland"},
			StreetAddress: []string{"Adelaide Street"},
			PostalCode:    []string{"4000"},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("generate ca pk: %w", err)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, fmt.Errorf("generate ca cert: %w", err)
	}
	if err = writeCertificate(certBytes, "target/ca.crt"); err != nil {
		return nil, fmt.Errorf("write ca cert: %w", err)
	}
	if err = writePrivateKey(privKey, "target/ca.key"); err != nil {
		return nil, fmt.Errorf("write ca key: %w", err)
	}
	return &rootCA{cert: cert, key: privKey}, nil
}

func (r *rootCA) createCert(alias string, now time.Time) error {
	log.Printf("generating %s cert\n", alias)
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2099),
		Subject: pkix.Name{
			CommonName:    fmt.Sprintf("%s.example.com", alias),
			Country:       []string{"AU"},
			Organization:  []string{"example grpc"},
			Locality:      []string{"Brisbane"},
			Province:      []string{"Queensland"},
			StreetAddress: []string{"Adelaide Street"},
			PostalCode:    []string{"4000"},
		},
		DNSNames:    []string{fmt.Sprintf("%s.example.com", alias), "localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   now,
		NotAfter:    now.AddDate(0, 1, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("generate %s pk: %w", alias, err)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, r.cert, &privKey.PublicKey, r.key)
	if err != nil {
		return fmt.Errorf("generate %s cert: %w", alias, err)
	}
	if err = writeCertificate(certBytes, fmt.Sprintf("target/%s.crt", alias)); err != nil {
		return fmt.Errorf("write %s cert: %w", alias, err)
	}
	if err = writePrivateKey(privKey, fmt.Sprintf("target/%s.key", alias)); err != nil {
		return fmt.Errorf("write %s key: %w", alias, err)
	}
	return nil
}

func writeCertificate(certBytes []byte, outfile string) error {
	fp, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer fp.Close()

	return pem.Encode(fp, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
}

func writePrivateKey(pk *rsa.PrivateKey, outfile string) error {
	fp, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer fp.Close()

	return pem.Encode(fp, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	})
}
