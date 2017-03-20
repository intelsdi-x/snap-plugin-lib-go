// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

// certTestUtil offers a few methods to generate a few self-signed certificates
// suitable only for test.
type certTestUtil struct {
}

const (
	keyBitsDefault            = 2048
	defaultKeyValidPeriod     = 6 * time.Hour
	rsaKeyPEMHeader           = "RSA PRIVATE KEY"
	certificatePEMHeader      = "CERTIFICATE"
	defaultSignatureAlgorithm = x509.SHA256WithRSA
	defaultPublicKeyAlgorithm = x509.RSA
)

func (u certTestUtil) writePEMFile(fn string, pemHeader string, b []byte) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	pem.Encode(w, &pem.Block{
		Type:  pemHeader,
		Bytes: b,
	})
	w.Flush()
	return nil
}

func (u certTestUtil) makeCACertKeyPair(caName, ouName string, keyValidPeriod time.Duration) (caCertTpl *x509.Certificate, caCertBytes []byte, caPrivKey *rsa.PrivateKey, err error) {
	caPrivKey, err = rsa.GenerateKey(rand.Reader, keyBitsDefault)
	if err != nil {
		return nil, nil, nil, err
	}
	caPubKey := caPrivKey.Public()
	caPubBytes, err := x509.MarshalPKIXPublicKey(caPubKey)
	if err != nil {
		return nil, nil, nil, err
	}
	caPubSha256 := sha256.Sum256(caPubBytes)
	caCertTpl = &x509.Certificate{
		SignatureAlgorithm: defaultSignatureAlgorithm,
		PublicKeyAlgorithm: defaultPublicKeyAlgorithm,
		Version:            3,
		SerialNumber:       big.NewInt(1),
		Subject: pkix.Name{
			CommonName: caName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(keyValidPeriod),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		MaxPathLenZero:        true,
		IsCA:                  true,
		SubjectKeyId:          caPubSha256[:],
	}
	caCertBytes, err = x509.CreateCertificate(rand.Reader, caCertTpl, caCertTpl, caPubKey, caPrivKey)
	if err != nil {
		return nil, nil, nil, err
	}
	return caCertTpl, caCertBytes, caPrivKey, nil
}

func (u certTestUtil) makeSubjCertKeyPair(cn, ou string, keyValidPeriod time.Duration, caCertTpl *x509.Certificate, caPrivKey *rsa.PrivateKey) (subjCertBytes []byte, subjPrivKey *rsa.PrivateKey, err error) {
	subjPrivKey, err = rsa.GenerateKey(rand.Reader, keyBitsDefault)
	if err != nil {
		return nil, nil, err
	}
	subjPubBytes, err := x509.MarshalPKIXPublicKey(subjPrivKey.Public())
	if err != nil {
		return nil, nil, err
	}
	subjPubSha256 := sha256.Sum256(subjPubBytes)
	subjCertTpl := x509.Certificate{
		SignatureAlgorithm: defaultSignatureAlgorithm,
		PublicKeyAlgorithm: defaultPublicKeyAlgorithm,
		Version:            3,
		SerialNumber:       big.NewInt(1),
		Subject: pkix.Name{
			OrganizationalUnit: []string{ou},
			CommonName:         cn,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(keyValidPeriod),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyAgreement,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		SubjectKeyId: subjPubSha256[:],
	}
	subjCertTpl.DNSNames = strings.Fields("localhost")
	subjCertTpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	subjCertBytes, err = x509.CreateCertificate(rand.Reader, &subjCertTpl, caCertTpl, subjPrivKey.Public(), caPrivKey)
	return subjCertBytes, subjPrivKey, err
}
