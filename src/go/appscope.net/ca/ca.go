package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"appscope.net/api/cert/v1"
)

const (
	kRootCAFile     = "root-ca.crt"
	kTlsCAKeyFile   = "tls-ca.key"
	kTlsCAFile      = "tls-ca.crt"
	kVpnCertKeyFile = "vpn-cert.key"
	kVpnCertFile    = "vpn-cert.crt"
)

var RootCA tls.Certificate
var TlsCA tls.Certificate
var VpnCert *x509.Certificate
var VpnKey *rsa.PrivateKey
var RootCAUrl string
var TlsCAUrl string

func RequestCertificates(dir string) error {

	httpClient := &http.Client{}
	certSvc, err := cert.New(httpClient)
	if err != nil {
		return err
	}

	VpnKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}

	req := new(cert.SignVPNCertRequest)
	req.PublicKey, err = x509.MarshalPKIXPublicKey(&VpnKey.PublicKey)
	if err != nil {
		return err
	}

	resp, err := certSvc.Signvpncert(req).Do()
	if err != nil {
		return err
	} else if resp.Error != nil {
		return fmt.Errorf("Failed to sign %+v : %v", req, resp.Error)
	}

	RootCAUrl = resp.RootCAUrl
	TlsCAUrl = resp.TlsCAUrl

	RootCA = tls.Certificate{
		Certificate: [][]byte{resp.RootCA},
	}

	RootCA.Leaf, err = x509.ParseCertificate(resp.RootCA)
	if err != nil {
		return err
	}

	TlsCA = tls.Certificate{
		Certificate: [][]byte{resp.TlsCA, resp.RootCA},
	}

	TlsCA.PrivateKey, err = x509.ParsePKCS1PrivateKey(resp.TlsCAKey)
	if err != nil {
		return err
	}

	TlsCA.Leaf, err = x509.ParseCertificate(resp.TlsCA)
	if err != nil {
		return err
	}

	VpnCert, err = x509.ParseCertificate(resp.VpnCert)
	if err != nil {
		return err
	}

	if err = checkCertificates(); err != nil {
		return err
	}

	if len(dir) > 0 {
		if err = storeCertificates(dir); err != nil {
			return err
		}
	}

	return nil
}

func storeCertificates(dir string) error {

	var err error

	if err = save2file(path.Join(dir, kRootCAFile), RootCA.Certificate[0], "CERTIFICATE"); err != nil {
		return err
	}

	privKey, ok := TlsCA.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return errors.New("Invalid TLS private key type.")
	}

	if err = save2file(path.Join(dir, kTlsCAKeyFile), x509.MarshalPKCS1PrivateKey(privKey), "RSA PRIVATE KEY"); err != nil {
		return err
	}

	if err = save2file(path.Join(dir, kTlsCAFile), TlsCA.Certificate[0], "CERTIFICATE"); err != nil {
		return err
	}

	if err = save2file(path.Join(dir, kVpnCertKeyFile), x509.MarshalPKCS1PrivateKey(VpnKey), "RSA PRIVATE KEY"); err != nil {
		return err
	}

	if err = save2file(path.Join(dir, kVpnCertFile), VpnCert.Raw, "CERTIFICATE"); err != nil {
		return err
	}

	return nil
}

func LoadCertificates(dir string) error {

	bytes, err := ioutil.ReadFile(path.Join(dir, kRootCAFile))
	if err != nil {
		return err
	}

	block, _ := pem.Decode(bytes)
	if err != nil {
		return err
	}

	RootCA = tls.Certificate{
		Certificate: [][]byte{block.Bytes},
	}

	RootCA.Leaf, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	TlsCA, err = tls.LoadX509KeyPair(path.Join(dir, kTlsCAFile), path.Join(dir, kTlsCAKeyFile))
	if err != nil {
		return err
	}

	TlsCA.Certificate = append(TlsCA.Certificate, block.Bytes)
	TlsCA.Leaf, err = x509.ParseCertificate(TlsCA.Certificate[0])
	if err != nil {
		return err
	}

	bytes, err = readFromFile(path.Join(dir, kVpnCertFile))
	if err != nil {
		return err
	}

	VpnCert, err = x509.ParseCertificate(bytes)
	if err != nil {
		return err
	}

	bytes, err = readFromFile(path.Join(dir, kVpnCertKeyFile))
	if err != nil {
		return err
	}

	VpnKey, err = x509.ParsePKCS1PrivateKey(bytes)
	if err != nil {
		return err
	}

	if err = checkCertificates(); err != nil {
		return err
	}

	return nil
}

func checkCertificates() error {

	now := time.Now()

	if RootCA.Leaf.NotAfter.Before(now) {
		return fmt.Errorf("Root CA certificate has expired.")
	}

	if TlsCA.Leaf.NotAfter.Before(now) {
		return fmt.Errorf("TLS CA certificate has expired.")
	}

	if VpnCert.NotAfter.Before(now) {
		return fmt.Errorf("VPN certificate has expired.")
	}

	return nil
}

func save2file(path string, bytes []byte, typ string) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	block := &pem.Block{
		Type:  typ,
		Bytes: bytes,
	}
	if err = pem.Encode(f, block); err != nil {
		return err
	}

	return nil
}

func readFromFile(path string) ([]byte, error) {

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, fmt.Errorf("PEM file \"%s\" is empty.", path)
	}

	return block.Bytes, nil
}
