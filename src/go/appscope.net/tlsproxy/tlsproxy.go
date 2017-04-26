package tlsproxy

import (
	"appscope.net/ca"
	"appscope.net/capture"
	"appscope.net/conntrac"
	"appscope.net/filter"
	"appscope.net/util"

	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/google/gopacket/layers"
)

var tlsClientSkipVerify = &tls.Config{InsecureSkipVerify: true}

type TLSProxy struct {
	LocalAddr, RemoteAddr *string
	Tracker               *capture.ConnectionTracker
	FilterController      *filter.Controller

	StaticExclByDomain   map[string]bool
	exclByClientByDomain map[string]map[string]string // [device Id, [domain, error]]
	exclDomainsMutex     sync.RWMutex
	EnableFiltering      bool
}

func (pxy *TLSProxy) Listen(dir string) error {

	var err error = nil
	if len(dir) > 0 {
		if err = ca.LoadCertificates(dir); err != nil {
			glog.Errorf("Error loading signing cert & key: %v", err)
		}
	}
	if len(dir) == 0 || err != nil {
		if err = ca.RequestCertificates(dir); err != nil {
			glog.Errorf("Error requesting signing cert & key: %v", err)
			return err
		}
	}

	certificateCache.entries = make(map[string][]tls.Certificate)

	addr, err := net.ResolveTCPAddr("tcp", *pxy.LocalAddr)
	if err != nil {
		glog.Errorf("Resolve local address error: %v", err)
		return err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		glog.Errorf("Listen %v error: %v", *pxy.LocalAddr, err)
		return err
	}

	glog.Infof("TLS HTTP Proxy enable filtering = %v", pxy.EnableFiltering)

	glog.Infof("TLS HTTP Proxy listening at %s", *pxy.LocalAddr)
	for {
		lConn, err := listener.AcceptTCP()
		if err == nil {
			go util.SafeRun(func() { pxy.connect(lConn) })
		} else {
			glog.Errorf("AcceptTCP() error %v", err)
		}
	}
}

func (pxy *TLSProxy) connect(lConn *net.TCPConn) {
	tm := time.Now()
	
	defer lConn.Close()
	
	connTrackRec := &conntrac.ConntracRec{
		Timestamp: uint64(time.Now().Unix()),
		ConnStartTime: &tm,
		TlsConn  : true,
	}

	lConn, rAddr, serverName := GetOriginalDst(lConn, TlsServerNameLookup)
	connTrackRec.ClientIP = layers.NewIPEndpoint((lConn.RemoteAddr()).(*net.TCPAddr).IP.To4())
	connTrackRec.ServerName = serverName
	connTrackRec.ServerAddr = rAddr.String()
	
	glog.Infof("OriginalDst(%v) = %v", rAddr, serverName)
	if rAddr == nil {
		registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_GET_ORIGINAL_DST_ERR)
		
		glog.Errorf("Failed to get original destination for %+v", lConn)		
		return
	}

	rConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_RDIAL_TCP_ERR)
		
		glog.Errorf("Connect with remote host error %v", err)
		return
	}
	connTrackRec.ServerIP = layers.NewIPEndpoint((rConn.RemoteAddr()).(*net.TCPAddr).IP.To4())

	defer rConn.Close()

	if pxy.useMitm(serverName, lConn, rConn) {
		pxy.mitmConnect(serverName, lConn, rConn, connTrackRec)
	} else {
		registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_MITM_EXCLUDED_ERR)
		
		connectNoRecording(lConn, rConn)
	}
}

func (pxy *TLSProxy) mitmConnect(serverName string, lConn *net.TCPConn, rConn *net.TCPConn, connTrackRec *conntrac.ConntracRec) {
	// --------- handshake with server --------------------------
	tServerStart := time.Now()

	var err error
	var dServerHandshake time.Duration
	var rTlsConn *tls.Conn
	var config *tls.Config = &tls.Config{
		ClientAuth:         tls.NoClientCert,
		InsecureSkipVerify: true}

	glog.Infof("Server name '%v'", serverName)
	
	if serverName == "" {
		rTlsConn = tls.Client(rConn, &tls.Config{InsecureSkipVerify: true})
		defer rTlsConn.Close()

		if err = rTlsConn.Handshake(); err != nil {
			glog.Errorf("TLS connect to %v failed : %v", rConn.RemoteAddr(), err)
			dumpCerts(rTlsConn.ConnectionState().PeerCertificates)
			pxy.exclDomainsAdd(serverName, lConn, rConn, err)
			
			registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_RTLS_HANDSHAKE_ERR)			
			return
		}
		dServerHandshake = time.Since(tServerStart)
		if dnsNames := getDomainNames(rTlsConn.ConnectionState().PeerCertificates); dnsNames == nil {
			err := errors.New(fmt.Sprintf("Server %v did not provide any certificate chains", rConn.RemoteAddr()))
			glog.Errorf(err.Error())
			pxy.exclDomainsAdd(serverName, lConn, rConn, err)
			
			registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_NO_CERT_PROVIDED_ERR)
			return
		} else if hostCertChain, err := getHostCertByNames(dnsNames); err != nil {
			glog.Errorf("Failed to generate certs for %v", dnsNames)
			pxy.exclDomainsAdd(serverName, lConn, rConn, err)
			
			registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_GENERATE_CERT_ERR)
			return
		} else {
			config.Certificates = hostCertChain
			glog.Infof("Signed for derived hostnames %v", dnsNames)
		}
	} else {
		rTlsConn = tls.Client(rConn, &tls.Config{ServerName: serverName, InsecureSkipVerify: true})

		if err = rTlsConn.Handshake(); err != nil {
			glog.Errorf("MITM handshake error with server %v:%s error=%v\n", rConn, serverName, err)
			pxy.exclDomainsAdd(serverName, lConn, rConn, err)
			
			registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_RTLS_HANDSHAKE_ERR)
			return
		} else if hostCertChain, err := getHostCertByName(serverName); err != nil {
			glog.Errorf("Failed to sign for %s : %v\n", serverName, err)
			pxy.exclDomainsAdd(serverName, lConn, rConn, err)
			
			registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_FAILED_TO_SIGN_ERR)
			return
		} else {
			dServerHandshake = time.Since(tServerStart)
			config.Certificates = hostCertChain
		}
	}
	
	connTrackRec.ServerTlsHandshakeStart    = &tServerStart
	connTrackRec.ServerTlsHandshakeDuration = dServerHandshake

	// --------- handshake with originating client --------------
	tClientStart := time.Now()
	lTlsConn := tls.Server(lConn, config)
	defer lTlsConn.Close()

	if err = lTlsConn.Handshake(); err != nil {
		glog.Errorf("MITM client handshake for %s error: %v", serverName, err)
		pxy.exclDomainsAdd(serverName, lConn, rConn, err)
		
		registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_LTLS_HANDSHAKE_ERR)
		return
	}
	
	if lTlsConn != nil {
		clientStatus := lTlsConn.ConnectionState()
		glog.Infof("MITM client asked for server=%s, handshake complete=%v", clientStatus.ServerName, clientStatus.HandshakeComplete)
		dumpCerts(clientStatus.PeerCertificates)
	}

	dClientHandshake := time.Since(tClientStart)

	if glog.V(vConnectionTrace) {
		glog.Infof("MITM TLS %s : %v<>%v handshake: %s, %s", serverName,
			lConn.RemoteAddr(), rConn.RemoteAddr(),
			dClientHandshake, dServerHandshake)
	}
	
	connTrackRec.ClientTlsHandshakeStart    = &tClientStart
	connTrackRec.ClientTlsHandshakeDuration = dClientHandshake
		
	if rTlsConn != nil {
		connTrackRec.ServerTlsVersion = rTlsConn.ConnectionState().Version
		connTrackRec.ServerTlsCipher  = rTlsConn.ConnectionState().CipherSuite
		connTrackRec.ServerTlsProto   = rTlsConn.ConnectionState().NegotiatedProtocol
	}
	if lTlsConn != nil {
		connTrackRec.ClientTlsVersion = lTlsConn.ConnectionState().Version
		connTrackRec.ClientTlsCipher  = lTlsConn.ConnectionState().CipherSuite
		connTrackRec.ClientTlsProto   = lTlsConn.ConnectionState().NegotiatedProtocol
	}

	if pxy.EnableFiltering && (pxy.FilterController != nil) {
		modifier := pxy.FilterController.CreateModifier((lConn.RemoteAddr()).(*net.TCPAddr).IP.To4().String(),
			layers.NewIPEndpoint((lConn.RemoteAddr()).(*net.TCPAddr).IP.To4()),
			lTlsConn, rTlsConn, pxy.Tracker)
                
		connectModifierAndRecord(pxy.Tracker, serverName, connTrackRec, lTlsConn, rTlsConn, modifier)

                pxy.FilterController.DeleteModifier(modifier)
	} else {
		connectAndRecord(pxy.Tracker, serverName, connTrackRec, lTlsConn, rTlsConn)
        }
        
        tm := time.Now()
        connTrackRec.ConnEndTime = &tm
}

var transparentTimeout = time.Minute * 10

func (pxy *TLSProxy) exclDomainsSet(clientId string, domains map[string]string) {	
	pxy.exclDomainsMutex.Lock()
	defer pxy.exclDomainsMutex.Unlock()
	
	pxy.exclByClientByDomain[clientId] = domains
}

func (pxy *TLSProxy) exclDomainsGet(clientId string) (domains map[string]string) {
	pxy.exclDomainsMutex.RLock()
	defer pxy.exclDomainsMutex.RUnlock()
	
	if exclByDomain, there := pxy.exclByClientByDomain[clientId]; there {
		domains = exclByDomain
	}
	
	return domains
}

func (pxy *TLSProxy) exclDomainsAdd(serverName string, lConn *net.TCPConn, rConn *net.TCPConn, err error) {
	var key string
	
	clientId := lConn.RemoteAddr().(*net.TCPAddr).IP.To4().String()
	
	if (serverName != "") {
		key = serverName
	} else {
		key = rConn.RemoteAddr().(*net.TCPAddr).IP.To4().String()
	}
	
	glog.Infof("Add domain to exclusion: clientId: %v, domain %v, err %v", clientId, key, err)
	
	pxy.exclDomainsMutex.Lock()
	defer pxy.exclDomainsMutex.Unlock()
	
	var exclByDomain map[string]string
	var there bool

	if exclByDomain, there = pxy.exclByClientByDomain[clientId]; !there {
		exclByDomain = make(map[string]string)
		pxy.exclByClientByDomain[clientId] = exclByDomain
	}
	
	exclByDomain[key] = err.Error()
}

func (pxy *TLSProxy) useMitm(serverName string, lConn *net.TCPConn, rConn *net.TCPConn) bool {
	var key string
	
	if (serverName != "") {
		key = serverName
	} else {
		key = rConn.RemoteAddr().(*net.TCPAddr).IP.To4().String()
	}
	
	if _, there := pxy.StaticExclByDomain[key]; there {
		glog.Infof("MITM: excluded domain '%v'", key)
		return false
	}
	
	pxy.exclDomainsMutex.RLock()
	defer pxy.exclDomainsMutex.RUnlock()
	
	if exclByDomain, there := pxy.exclByClientByDomain[lConn.RemoteAddr().(*net.TCPAddr).IP.To4().String()]; there {
		if _, there := exclByDomain[key]; there {
			glog.Infof("MITM: excluded domain '%v'", key)
			return false
		}
	}
	
	return true
}
