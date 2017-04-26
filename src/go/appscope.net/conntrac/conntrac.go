package conntrac

/*
 this module provides mapping between observed on outbound network interface host:port combinations
 into before-NAT (real) ones

 It only properly works on Linux 64-bit
 OSX is dummy implementation

*/

import (
	"fmt"
	"github.com/google/gopacket"
	"strconv"
	"time"
)

type ConntrackTcpAttr uint8

// see http://www.netfilter.org/projects/libnetfilter_conntrack/doxygen/libnetfilter__conntrack__tcp_8h_source.html

const (
	TCP_CONNTRACK_NONE        ConntrackTcpAttr = 0
	TCP_CONNTRACK_SYN_SENT    ConntrackTcpAttr = 1
	TCP_CONNTRACK_SYN_RECV    ConntrackTcpAttr = 2
	TCP_CONNTRACK_ESTABLISHED ConntrackTcpAttr = 3
	TCP_CONNTRACK_FIN_WAIT    ConntrackTcpAttr = 4
	TCP_CONNTRACK_CLOSE_WAIT  ConntrackTcpAttr = 5
	TCP_CONNTRACK_LAST_ACK    ConntrackTcpAttr = 6
	TCP_CONNTRACK_TIME_WAIT   ConntrackTcpAttr = 7
	TCP_CONNTRACK_CLOSE       ConntrackTcpAttr = 8
	TCP_CONNTRACK_SYN_SENT2   ConntrackTcpAttr = 9
)

type SegmentError uint32

const (
	SEGMENT_NO_ERR               SegmentError = 0
	SEGMENT_GET_ORIGINAL_DST_ERR SegmentError = 1
	SEGMENT_RDIAL_TCP_ERR        SegmentError = 2
	SEGMENT_RTLS_HANDSHAKE_ERR   SegmentError = 3
	SEGMENT_LTLS_HANDSHAKE_ERR   SegmentError = 4
	SEGMENT_NO_CERT_PROVIDED_ERR SegmentError = 5
	SEGMENT_GENERATE_CERT_ERR    SegmentError = 6
	SEGMENT_FAILED_TO_SIGN_ERR   SegmentError = 7
	SEGMENT_MITM_EXCLUDED_ERR    SegmentError = 8
)

type ConntracRec struct {
	ClientIP, ClientPort, ServerIP, ServerPort, NatIP, NatPort gopacket.Endpoint
	Timestamp                   uint64
	TcpState                    ConntrackTcpAttr
	ConnErr                     SegmentError
	ConnStartTime               *time.Time
	ConnEndTime                 *time.Time
	TlsConn                     bool
	ServerName                  string
	ServerAddr                  string
	ServerTlsHandshakeStart     *time.Time
	ServerTlsHandshakeDuration  time.Duration
	ServerTlsVersion            uint16
	ServerTlsCipher             uint16
	ServerTlsProto              string
	ClientTlsHandshakeStart     *time.Time
	ClientTlsHandshakeDuration  time.Duration
	ClientTlsVersion            uint16
	ClientTlsCipher             uint16
	ClientTlsProto              string
}

func (cr *ConntracRec) State() string {
	return tcpstate2string(cr.TcpState)
}

func (cr *ConntracRec) String() string {
	return fmt.Sprintf("[%s]\t %d cli=%v:%v srv=%v:%v nat=%v:%v err=%v",
		tcpstate2string(cr.TcpState),
		cr.Timestamp,
		cr.ClientIP, cr.ClientPort,
		cr.ServerIP, cr.ServerPort,
		cr.NatIP, cr.NatPort, cr.ConnErr)
}

func tcpstate2string(state ConntrackTcpAttr) string {
	if val, there := tcpstateStr[state]; there {
		return val
	} else {
		return strconv.FormatUint(uint64(state), 2)
	}
}

var tcpstateStr = map[ConntrackTcpAttr]string{
	TCP_CONNTRACK_NONE:        "NONE",
	TCP_CONNTRACK_SYN_SENT:    "SYN_SENT",
	TCP_CONNTRACK_SYN_RECV:    "SYN_RECV",
	TCP_CONNTRACK_ESTABLISHED: "ESTABLISHED",
	TCP_CONNTRACK_FIN_WAIT:    "FIN_WAIT",
	TCP_CONNTRACK_CLOSE_WAIT:  "CLOSE_WAIT",
	TCP_CONNTRACK_LAST_ACK:    "LAST_ACK",
	TCP_CONNTRACK_TIME_WAIT:   "TIME_WAIT",
	TCP_CONNTRACK_CLOSE:       "CLOSE",
	TCP_CONNTRACK_SYN_SENT2:   "SYN_SENT2",
}
