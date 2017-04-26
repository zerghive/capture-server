package tlsproxy

import (
	"crypto/tls"
	"io"
	"net"
	"syscall"

	"github.com/golang/glog"
)

const (
	NoTlsServerNameLookup = false
	TlsServerNameLookup   = true

	SO_ORIGINAL_DST = 80

	tlsHeaderLen                = 5
	tlsContentTypeHandshake     = 22
	tlsHandshakeTypeClientHello = 1
	tlsExtensionServerName      = 0
	tlsServerNameHost           = 0
)

func GetOriginalDst(conn *net.TCPConn, tlsLookup bool) (newConn *net.TCPConn, rAddr *net.TCPAddr, rHost string) {

	rAddr = nil
	newConn = conn

	// File sets the underlying os.File to blocking mode and returns a copy.
	file, err := conn.File()
	if err == nil {

		defer func() { file.Close() }()

		fd := int(file.Fd())

		addr, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
		if err == nil {
			rAddr = &net.TCPAddr{
				IP:   net.IP{addr.Multiaddr[4], addr.Multiaddr[5], addr.Multiaddr[6], addr.Multiaddr[7]},
				Port: int(addr.Multiaddr[2])<<8 + int(addr.Multiaddr[3]),
			}
			//fmt.Printf("[!] Original dst: %v.%v.%v.%v:%v\n", rAddr.IP[0],
			//	rAddr.IP[1], rAddr.IP[2], rAddr.IP[3], rAddr.Port)
		} else {
			glog.Errorf("[!] Original dst address not found:", err)
		}

		if tlsLookup {
			rHost = getServerName(fd)
		} else {
			rHost = rAddr.String()
		}

		// Create a new TCPConn. The new TCPConn will be in non-blocking mode.
		fileConn, err := net.FileConn(file)
		if err == nil {
			defer func() { conn.Close() }()
			newConn = fileConn.(*net.TCPConn)
		} else {
			glog.Errorf("[!] File->Socket conversion error:", err)
		}
	}

	return newConn, rAddr, rHost
}

func getServerName(fd int) string {

	// Read TLS header.
	data, err := peek(fd, tlsHeaderLen)
	if err != nil {
		return ""
	}

	typ := int(data[0])
	ver := int(data[1])<<8 | int(data[2])
	n := int(data[3])<<8 | int(data[4])
	// fmt.Printf("[+] TLS content type: %v, ver: 0x%X, len: %v.\n", typ, ver, n)

	if n < 4 || typ != tlsContentTypeHandshake || ver < tls.VersionTLS10 {
		return ""
	}

	// Read client hello packet.
	data, err = peek(fd, tlsHeaderLen+n)
	if err != nil {
		return ""
	}

	data = data[tlsHeaderLen:]
	typ = int(data[0])
	n = int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	if len(data) <= 38 || typ != tlsHandshakeTypeClientHello {
		return ""
	}

	// Session ID Length.
	n = int(data[38])
	if len(data) <= 39+n+2 {
		return ""
	}

	// Cipher Suite Length.
	data = data[39+n:]
	n = int(data[0])<<8 | int(data[1])

	if len(data) <= 2+n+1 {
		return ""
	}

	// Compression Methods Length.
	data = data[2+n:]
	n = int(data[0])
	if len(data) <= 1+n+2 {
		return ""
	}

	// Extensions Length.
	data = data[1+n:]
	n = int(data[0])<<8 | int(data[1])
	if len(data) != 2+n {
		return ""
	}

	// Process extensions. Extension header is equal to 4 bytes.
	data = data[2:]
	for len(data) >= 4 {

		typ = int(data[0])<<8 | int(data[1]) // Extension Type.
		n = int(data[2])<<8 | int(data[3])   // Extension Length.
		if len(data) < 4+n {
			break
		}

		data = data[4:]
		if typ != tlsExtensionServerName {
			data = data[n:]
			continue
		}

		// Process Server Name extension.
		if n < 2 {
			break
		}

		// Server Names List Length.
		listLength := int(data[0])<<8 | int(data[1])
		if len(data) < 2+listLength {
			break
		}

		// Process list items. Item header is equal to 3 bytes.
		data = data[2:]
		for listLength >= 3 {

			typ = int(data[0])                 // List Item Type.
			n = int(data[1])<<8 | int(data[2]) // List Item Length.
			if listLength < 3+n {
				break
			}

			data = data[3:]
			listLength -= 3 + n
			if typ == tlsServerNameHost {
				return string(data[0:n])
			}

			data = data[n:]
		}
	}

	return ""
}

func peek(fd int, n int) (data []byte, err error) {

	data = make([]byte, 0, n)

	for {
		m, _, err := syscall.Recvfrom(fd, data[len(data):cap(data)], syscall.MSG_PEEK|syscall.MSG_WAITALL)
		if err != nil {
			return nil, err
		} else if m == 0 {
			// man recvfrom(2) : The return value will be 0 when the peer has performed an orderly shutdown
			return nil, io.EOF
		}

		data = data[0 : len(data)+m]
		if len(data) >= n {
			return data, nil
		}
	}
}
