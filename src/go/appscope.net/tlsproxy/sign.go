package tlsproxy

import (
	"appscope.net/ca"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
)

var certificateCache struct {
	sync.RWMutex
	certPriv *rsa.PrivateKey
	entries  map[string][]tls.Certificate
}

func init() {
	var err error
	if certificateCache.certPriv, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		panic("Failed to generate key")
	}
}

func hashSorted(lst []string) []byte {
	c := make([]string, len(lst))
	copy(c, lst)
	sort.Strings(c)
	h := sha1.New()
	for _, s := range c {
		h.Write([]byte(s + ","))
	}
	return h.Sum(nil)
}

func hashSortedBigInt(lst []string) *big.Int {
	rv := new(big.Int)
	rv.SetBytes(hashSorted(lst))
	return rv
}

var signerVersion = ":mitm"

func getHostCertByName(host string) ([]tls.Certificate, error) {
	return getHostCert([]string{host}, host)
}

func getHostCertByNames(hosts []string) ([]tls.Certificate, error) {
	return getHostCert(hosts, strings.Join(hosts, ","))
}

var (
	certStartTime = time.Now().Add(-time.Hour * 24 * 365)
	certEndTime   = time.Now().Add(time.Hour * 24 * 365)
)

func getHostCert(hosts []string, key string) ([]tls.Certificate, error) {
	certificateCache.RLock()
	certificates, there := certificateCache.entries[key]
	certificateCache.RUnlock()

	if there {
		return certificates, nil
	}

	certificateCache.Lock()
	defer certificateCache.Unlock()

	if hostCert, err := signHost(hosts); err == nil {
		certificateCache.entries[key] = []tls.Certificate{hostCert}
		return certificateCache.entries[key], nil
	} else {
		return nil, err
	}
}

func signHost(hosts []string) (cert tls.Certificate, err error) {

	hash := hashSorted(append(hosts, signerVersion, ":"+runtime.Version()))
	serial := new(big.Int)
	serial.SetBytes(hash)
	template := x509.Certificate{
		SerialNumber: serial,
		Issuer:       ca.TlsCA.Leaf.Subject,
		Subject: pkix.Name{
			Organization: ca.TlsCA.Leaf.Subject.Organization,
		},
		NotBefore: certStartTime,
		NotAfter:  certEndTime,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: false,

		IssuingCertificateURL: []string{
			ca.RootCAUrl,
		},

		AuthorityKeyId: ca.TlsCA.Leaf.SubjectKeyId,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}
	var derBytes []byte
	if derBytes, err = x509.CreateCertificate(rand.Reader, &template, ca.TlsCA.Leaf,
		&certificateCache.certPriv.PublicKey, ca.TlsCA.PrivateKey); err != nil {
		return
	}
	return tls.Certificate{
		Certificate: [][]byte{derBytes, ca.TlsCA.Certificate[0], ca.RootCA.Certificate[0]},
		PrivateKey:  certificateCache.certPriv,
	}, nil
}

func dumpCerts(certs []*x509.Certificate) {
	for _, c := range certs {
		glog.Infof("   Subject: %+v, DNS: %+v, IP: %+v, PermittedDNSDomains", c.Subject, c.DNSNames, c.IPAddresses, c.PermittedDNSDomains)
	}
}

func getDomainNames(certs []*x509.Certificate) []string {
	for _, c := range certs {
		if len(c.DNSNames) > 0 {
			return c.DNSNames
		}
	}
	return nil
}
