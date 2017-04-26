package conntrac

/*
#include <stdlib.h>
#include <arpa/inet.h>
#include <linux/netlink.h>
#include <linux/netfilter/nfnetlink.h>
#include <linux/netfilter/nfnetlink_conntrack.h>
*/
import "C"

import (
	"github.com/golang/glog"

	nfct "github.com/chamaken/cgolmnfct"
	mnl "github.com/chamaken/cgolmnl"

	"appscope.net/util"

	"fmt"
	"os"
	"runtime"
	"syscall"
)

const (
	vConntracDebug = 10
)

/*
Example detecting NAT-ed connection, we would observe the following for client 10.42.42.1 looking for HTTPS(443) 198.41.191.47

[NEW]    tcp 6 120    SYN_SENT    src=10.42.42.1 dst=198.41.191.47 sport=58565 dport=443 [UNREPLIED] src=198.41.191.47 dst=62.210.220.75 sport=443 dport=58565
[UPDATE] tcp 6 60     SYN_RECV    src=10.42.42.1 dst=198.41.191.47 sport=58565 dport=443 src=198.41.191.47 dst=62.210.220.75 sport=443 dport=58565
[UPDATE] tcp 6 432000 ESTABLISHED src=10.42.42.1 dst=198.41.191.47 sport=58565 dport=443 src=198.41.191.47 dst=62.210.220.75 sport=443 dport=58565 [ASSURED]

Resulting connection tracker would be :
 ClientIP  : ATTR_ORIG_IPV4_SRC
 NatIP     : ATTR_REPL_IPV4_DST
 NatPort   : ATTR_REPL_PORT_DST
 Server    : ATTR_ORIG_IPV4_DST
 ServerPort: ATTR_REPL_PORT_SRC

*/

type conntrac_s struct {
	result util.Queue
}

func (track *conntrac_s) data_cb(nlh *mnl.Nlmsghdr, data interface{}) (int, syscall.Errno) {
	var msg_type nfct.ConntrackMsgType

	switch nlh.Type & 0xFF {
	case C.IPCTNL_MSG_CT_NEW:
		if nlh.Flags&(C.NLM_F_CREATE|C.NLM_F_EXCL) != 0 {
			msg_type = nfct.NFCT_T_NEW
		} else {
			msg_type = nfct.NFCT_T_UPDATE
		}
	case C.IPCTNL_MSG_CT_DELETE:
		msg_type = nfct.NFCT_T_DESTROY
	}

	/*
		if msg_type != nfct.NFCT_T_NEW {
			return mnl.MNL_CB_OK, 0
		}
	*/

	ct, err := nfct.NewConntrack()
	if err != nil {
		return mnl.MNL_CB_OK, 0
	}
	defer ct.Destroy()
	ct.NlmsgParse(nlh)

	tcpOpt, _ := ct.AttrU8(nfct.ATTR_TCP_STATE)
	if ConntrackTcpAttr(tcpOpt) == TCP_CONNTRACK_NONE { // only interested in TCP
		return mnl.MNL_CB_OK, 0
	}

	clientIPv4, _ := ct.AttrU32(nfct.ATTR_ORIG_IPV4_SRC)
	clientPort, _ := ct.AttrU16(nfct.ATTR_ORIG_PORT_SRC)
	natIPv4, _ := ct.AttrU32(nfct.ATTR_REPL_IPV4_DST)
	natPort, _ := ct.AttrU16(nfct.ATTR_REPL_PORT_DST)
	serverIPv4, _ := ct.AttrU32(nfct.ATTR_ORIG_IPV4_DST)
	serverPort, _ := ct.AttrU16(nfct.ATTR_ORIG_PORT_DST)

	tstamp, _ := ct.AttrU64(nfct.ATTR_TIMESTAMP_START)
	// if timestamp is zero then its not enabled for netfilter / conntrack
	// "$ echo 1 > /proc/sys/net/netfilter/nf_conntrack_timestamp"

	if glog.V(vConntracDebug) {
		buf := make([]byte, 4096)
		ct.Snprintf(buf, msg_type, nfct.NFCT_O_DEFAULT, nfct.NFCT_OF_TIMESTAMP)
		glog.Info(string(buf))
	}

	rc := &ConntracRec{
		ServerIP:   util.Uint32toIP(serverIPv4),
		ServerPort: util.NUint16toPort(serverPort),
		NatIP:      util.Uint32toIP(natIPv4),
		NatPort:    util.NUint16toPort(natPort),
		ClientIP:   util.Uint32toIP(clientIPv4),
		ClientPort: util.NUint16toPort(clientPort),
		TcpState:   ConntrackTcpAttr(tcpOpt),
		Timestamp:  tstamp,
	}
	track.result.Push(rc)

	// fmt.Printf("%s | %d %d | %s\n", rc.String(), natPort, serverPort, buf)

	return mnl.MNL_CB_OK, 0
}

func NewConntrac(q util.Queue) {
	ct := &conntrac_s{q}
	go util.SafeRun(ct.runloop)
}

func (track *conntrac_s) runloop() {
	buf := make([]byte, mnl.MNL_SOCKET_BUFFER_SIZE)

	nl, err := mnl.NewSocket(C.NETLINK_NETFILTER)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mnl_socket_open: %s\n", err)
		os.Exit(C.EXIT_FAILURE)
	}
	defer nl.Close()

	if err = nl.Bind(C.NF_NETLINK_CONNTRACK_NEW|
		C.NF_NETLINK_CONNTRACK_UPDATE|
		C.NF_NETLINK_CONNTRACK_DESTROY,
		mnl.MNL_SOCKET_AUTOPID); err != nil {
		fmt.Fprintf(os.Stderr, "mnl_socket_bind: %s\n", err)
		os.Exit(C.EXIT_FAILURE)
	}

	runtime.LockOSThread()

	glog.Infof("Netfilter Conntrack running")
	ret := mnl.MNL_CB_OK
	for ret > 0 {
		nrecv, err := nl.Recvfrom(buf)
		if err != nil {
			glog.Errorf("mnl_socket_recvfrom: %s\n", err)
			continue
		}
		if ret, err = mnl.CbRun(buf[:nrecv], 0, 0, track.data_cb, nil); ret < 0 {
			glog.Errorf("mnl_cb_run: %s\n", err)
			continue
		}
	}
}
