package capture

/*
 listen to NAT connection tracking events, and perform matching between registering commands
 arriving from PF_RING and NAT / Netfilter_conntrack updates

 mapping is by observed port on a NAT'ed interface, that is :
 (pcap) client:port <==> (nat):port

*/

import (
	"appscope.net/event"
	"appscope.net/conntrac"

	"github.com/golang/glog"
	"github.com/google/gopacket"

	"fmt"
)

func (ct *ConnectionTracker) doRegisterStreamDuplex(v interface{}) (err error) {
	cmd := v.(*cmdRegisterStreamDuplex)

	glog.Infof("+ %v", cmd.track)
	
	conn := connection{
		requestStream:  cmd.request,
		responseStream: cmd.response,
		natRec:         cmd.track,
		connId:         cmd.connId,
	}
	
	clientInfo := ct.getClientInfo(cmd.track.ClientIP)
	
	clientInfo.connections.PushFront(&conn)
	clientInfo.connectionByConnId[cmd.connId] = &conn
	
	glog.Infof(" (added) ")

	return nil
}

func (ct *ConnectionTracker) doRegisterStreamLeg(v interface{}) (err error) {
	cmd := v.(*cmdRegisterStreamLeg)
	var port gopacket.Endpoint
	if cmd.streamType == Request {
		port = cmd.transport.Src()
	} else {
		port = cmd.transport.Dst()
	}

	rqrsp, there := ct.clients_pending[port]
	if !there {
		rqrsp = make([]*cmdRegisterStreamLeg, 2)
		ct.clients_pending[port] = rqrsp
	} else if rqrsp[cmd.streamType] != nil { // unexpected collision
		rqrsp[cmd.streamType].streamData.Close()
		rqrsp[cmd.streamType].streamData.Destroy()
		err = fmt.Errorf("Unexpected collision for port %v : overwriting %v with %v", port, rqrsp[cmd.streamType], cmd)
	}
	rqrsp[cmd.streamType] = cmd

	if glog.V(vConnectionParts) {
		glog.Infof("RQRSP[%s]=%v", port, rqrsp)
	}
	ct.tryAssociateConnections(port)
	return
}

/* takes request/response and NAT mapping and moves them into proper connection table */
func (ct *ConnectionTracker) tryAssociateConnections(port gopacket.Endpoint) {
	rqrsp := ct.clients_pending[port]
	natRec := ct.natPortmap[port]
	if glog.V(vConnectionParts) {
		glog.Infof("PORT[%v] nat=%v, rqrsp=%v", port, natRec, rqrsp)
	}
	if natRec == nil || rqrsp == nil {
		return
	}

	req := rqrsp[Request]
	resp := rqrsp[Response]
	if req == nil || resp == nil {
		return
	}

	/* i.e.
	   req= net=62.210.220.75->173.194.45.74, port=48790->80, stream=0xc2090e81b0,
	   resp=net=173.194.45.74->62.210.220.75, port=80->48790, stream=0xc2090e82d0,
	   nat=[SYN_SENT] 0 client=10.42.42.1 server=173.194.45.74:80 nat=62.210.220.75:48790
	*/
	// ensure we're mapping correct requests/responses
	if dont_match(req.net.Src(), resp.net.Dst(), natRec.NatIP) ||
		dont_match(req.net.Dst(), resp.net.Src(), natRec.ServerIP) ||
		dont_match(req.transport.Src(), resp.transport.Dst(), natRec.NatPort) ||
		dont_match(req.transport.Dst(), resp.transport.Src(), natRec.ServerPort) {
		glog.Errorf("req=%v, resp=%v, nat=%s don't match", req, resp, natRec)
		glog.Errorf("\n %v\n %v\n %v\n----\n %v\n %v\n %v",
			req.net.Src(), resp.net.Dst(), natRec.NatIP,
			req.net.Src().Raw(), resp.net.Dst().Raw(), natRec.NatIP.Raw())

		return
	}

	// cleanup entries
	rqrsp[Request] = nil
	rqrsp[Response] = nil
	ct.clients_pending[port] = nil
	ct.natPortmap[port] = nil

	// add new connection
	conn := &connection{
		natRec:         natRec,
		requestStream:  req.streamData,
		responseStream: resp.streamData,
	}

	ct.getClientInfo(natRec.ClientIP).connections.PushFront(conn)
	if glog.V(vConnectionTrace) {
		glog.Infof("%v+%v,%v,%v", natRec.ClientIP, req, resp, natRec)
	}

}

func (ct *ConnectionTracker) doNatConntracUpdate(v interface{}) (err error) {
	rec := v.(*conntrac.ConntracRec)

	ct.PostClientEvent(rec.ClientIP, func() event.Event {
		return &TcpConnEvent{
			Src:   fmt.Sprintf("%v:%v", rec.ClientIP, rec.ClientPort),
			Dst:   fmt.Sprintf("%v:%v", rec.ServerIP, rec.ServerPort),
			State: rec.State(),
		}
	})

	// TODO: enable back with port forwarding-compatible check
	// or maybe we don't need at all with port-forwarding
	return

	// ignore other interfaces
	// request from this machine
	// or request to this machine

	if (ct.natInterfaceAddress != rec.NatIP) ||
		(rec.ClientIP == rec.NatIP) ||
		(ct.natInterfaceAddress == rec.ServerIP) {
		if glog.V(vPacketTrace) {
			glog.Infof("Ignoring %v", rec)
		}
		return
	}

	// if this client never emitted http(s) request before, don't record his other connections
	// TODO : it will likely mean we might miss his very first request, but saves some cleanup time
	if clientEx := ct.clients[rec.ClientIP]; clientEx != nil {
		clientEx.conntracRecords.PushFront(rec)
	}

	// for request recording purposes, we're only interested in specific whitelabeled
	// ports and connection initiation events
	if rec.TcpState != conntrac.TCP_CONNTRACK_SYN_SENT ||
		ct.listenPorts[rec.ServerPort] == false {
		if glog.V(vPacketTrace) {
			glog.Infof("Discarding %v", rec)
		}
		return
	}

	if glog.V(vConnectionParts) {
		glog.Infof(" NAT[%s]=%v, was %v", rec.NatPort.String(), rec.String(), ct.natPortmap[rec.NatPort])
	}

	ct.natPortmap[rec.NatPort] = rec

	ct.tryAssociateConnections(rec.NatPort)
	return nil
}

func dont_match(one, two, three gopacket.Endpoint) bool {
	return !(one == two && one == three)
}
