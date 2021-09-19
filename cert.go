// Copyright 2021 The ssl-pairgen authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"syscall"
	"time"

	"golang.org/x/term"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

func (p *sslpair) makeCert() {
	priv, err := p.generateKey(false)
	fatalIfErr(err, "failed to generate user certificate key")
	pub := priv.(crypto.Signer).Public()

	expiration := time.Now().AddDate(1, 0, 0)

	tpl := &x509.Certificate{
		SerialNumber: randomSerialNumber(),
		Subject: pkix.Name{
			Organization: []string{p.orgName},
			CommonName:   p.userName,
		},

		NotBefore: time.Now(), NotAfter: expiration,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	cert, err := x509.CreateCertificate(rand.Reader, tpl, p.caCert, pub, p.caKey)
	fatalIfErr(err, "failed to generate certificate")

	certFile := p.userCertFile
	p12File := p.p12File

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	// privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	fatalIfErr(err, "failed to encode certificate key")
	// privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})

	err = ioutil.WriteFile(certFile, certPEM, 0644)
	fatalIfErr(err, "failed to save certificate")

	fmt.Printf("Enter passphrase for encrypting %s: ", p12File)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fatalIfErr(err, "failed to read passphrase")
	fmt.Println()
	password := string(bytePassword)
	fmt.Print("Re-enter passphrase: ")
	bytePassword, err = term.ReadPassword(int(syscall.Stdin))
	fatalIfErr(err, "failed to read passphrase")
	fmt.Println()
	check := string(bytePassword)
	if check != password {
		log.Fatalf("ERROR: password mismatch")
	}

	domainCert, _ := x509.ParseCertificate(cert)
	pfxData, err := pkcs12.Encode(rand.Reader, priv, domainCert, []*x509.Certificate{p.caCert}, password)
	fatalIfErr(err, "failed to generate PKCS#12")
	err = ioutil.WriteFile(p12File, pfxData, 0644)
	fatalIfErr(err, "failed to save PKCS#12")
	fmt.Printf("Wrote user private CA cert %s, browser package %s, and informational %s\n", p.certFile, p.p12File, p.userCertFile)
}

func (p *sslpair) generateKey(rootCA bool) (crypto.PrivateKey, error) {
	if rootCA {
		return rsa.GenerateKey(rand.Reader, 3072)
	}
	return rsa.GenerateKey(rand.Reader, 2048)
}

func randomSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	fatalIfErr(err, "failed to generate serial number")
	return serialNumber
}

func (p *sslpair) newCA() {
	priv, err := p.generateKey(true)
	fatalIfErr(err, "failed to generate the CA key")
	pub := priv.(crypto.Signer).Public()
	p.caKey = priv

	spkiASN1, err := x509.MarshalPKIXPublicKey(pub)
	fatalIfErr(err, "failed to encode public key")

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	fatalIfErr(err, "failed to decode public key")

	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)

	tpl := &x509.Certificate{
		SerialNumber: randomSerialNumber(),
		Subject: pkix.Name{
			Organization: []string{p.orgName},
			CommonName:   p.userName + "-privateCA",
		},
		SubjectKeyId: skid[:],

		NotAfter:  time.Now().AddDate(1, 0, 1),
		NotBefore: time.Now(),

		KeyUsage: x509.KeyUsageCertSign,

		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv)
	fatalIfErr(err, "failed to generate CA certificate")

	p.caCert, err = x509.ParseCertificate(cert)
	fatalIfErr(err, "failed to parse generated CA certificate")

	err = ioutil.WriteFile(p.certFile, pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: cert}), 0644)
	fatalIfErr(err, "failed to save CA cert")
}
