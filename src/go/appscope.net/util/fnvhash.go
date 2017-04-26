package util

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"encoding/binary"
	"net"
)

// See http://isthe.com/chongo/tech/comp/fnv/
func FnvHash(s []byte) (h uint64) {
	h = fnvBasis
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= fnvPrime
	}
	return
}

func FnvHash2(s1, s2 []byte) (h uint64) {
	h = fnvBasis
	for i := 0; i < len(s1); i++ {
		h ^= uint64(s1[i])
		h *= fnvPrime
	}
	for i := 0; i < len(s2); i++ {
		h ^= uint64(s2[i])
		h *= fnvPrime
	}

	return
}

const fnvBasis = 14695981039346656037
const fnvPrime = 1099511628211

func Uint32toIP(val uint32) gopacket.Endpoint {
	bb := make([]byte, 4)
	binary.LittleEndian.PutUint32(bb, val)
	return layers.NewIPEndpoint(net.IPv4(bb[0], bb[1], bb[2], bb[3]).To4())
}

// network byte order
func NUint16toPort(val uint16) gopacket.Endpoint {
	return gopacket.NewEndpoint(layers.EndpointTCPPort, []byte{byte(val), byte(val >> 8)})
}

// host byte order
func HUint16toPort(val uint16) gopacket.Endpoint {
	return gopacket.NewEndpoint(layers.EndpointTCPPort, []byte{byte(val >> 8), byte(val)})
}
