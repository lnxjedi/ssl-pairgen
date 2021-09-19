// Copyright 2021 The ssl-pairgen authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command mkcert is a simple zero-config tool to make development certificates.
package main

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"log"
	"os"
)

const usage = `Usage:

$ ssl-pairgen "<org>" <username>
    Generate <username>-ca.crt private CA, <username>.p12 browser package,
    and <username>.pem informational cert files.
`

type sslpair struct {
	certFile, userCertFile, p12File string
	orgName, userName               string
	caCert                          *x509.Certificate
	caKey                           crypto.PrivateKey
}

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Print(usage)
		return
	}
	orgName := args[0]
	userName := args[1]
	certFile := fmt.Sprintf("%s-ca.crt", userName)
	userCertFile := fmt.Sprintf("%s.pem", userName)
	p12File := fmt.Sprintf("%s.p12", userName)

	userpair := &sslpair{
		certFile:     certFile,
		userCertFile: userCertFile,
		p12File:      p12File,
		orgName:      orgName,
		userName:     userName,
	}
	userpair.generate()
}

func (p *sslpair) generate() {
	fmt.Println("Generating CA and user keys and certificates ...")
	p.newCA()
	p.makeCert()
}

func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err)
	}
}
