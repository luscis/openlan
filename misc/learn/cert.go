package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

func CertificateText(cert *x509.Certificate) (string, error) {
	var buf bytes.Buffer
	buf.Grow(4096) // 4KiB should be enough

	buf.WriteString(fmt.Sprintf("Certificate:\n"))
	buf.WriteString(fmt.Sprintf("%4sData:\n", ""))
	buf.WriteString(fmt.Sprintf("%8sSerial Number: %d (%#x)\n", "", cert.SerialNumber, cert.SerialNumber))
	buf.WriteString(fmt.Sprintf("%4sSignature Algorithm: %s\n", "", cert.SignatureAlgorithm))

	// Issuer information
	buf.WriteString(fmt.Sprintf("%8sIssuer: ", ""))
	// Validity information
	buf.WriteString(fmt.Sprintf("%8sValidity\n", ""))
	buf.WriteString(fmt.Sprintf("%12sNot Before: %s\n", "", cert.NotBefore.Format("Jan 2 15:04:05 2006 MST")))
	buf.WriteString(fmt.Sprintf("%12sNot After : %s\n", "", cert.NotAfter.Format("Jan 2 15:04:05 2006 MST")))

	now := time.Now()
	if now.Before(cert.NotBefore) {
		buf.WriteString(fmt.Sprintf("current time %s is before %s\n", now.Format(time.RFC3339), cert.NotBefore.Format(time.RFC3339)))

	} else if now.After(cert.NotAfter) {
		buf.WriteString(fmt.Sprintf("current time %s is after %s\n", now.Format(time.RFC3339), cert.NotAfter.Format(time.RFC3339)))
	}

	return buf.String(), nil
}

func main() {
	// Read and parse the PEM certificate file
	pemData, err := ioutil.ReadFile("cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	block, rest := pem.Decode(pemData)
	if block == nil || len(rest) > 0 {
		log.Fatal("Certificate decoding error")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	// Print the certificate
	result, err := CertificateText(cert)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(result)
}
