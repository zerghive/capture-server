package capture

import (
	"github.com/golang/glog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"appscope.net/conntrac"
	"appscope.net/event"
	"appscope.net/util"

	"container/list"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

/*
 Maintains a cache

 All commands are processed in command loop, this is how we avoid mutexes altogether
  - NAT connection mapping
  - stream registrations from packet capture
  - export commands take a current connection list and move it as a bulk for processing, and is inserted back once complete
  - data trimmer is the only process which modifies data concurrently (together with request / response writer process)
    that's the reason we take bunch of connections for export to separate scope, so they don't intersect

*/

const (
	Request  = 0
	Response = 1

	cChannelBuffer = 16

	cTrimDeadline = time.Minute
)

type connection struct {
	natRec                        *conntrac.ConntracRec
	requestStream, responseStream *SegmentedStream
	connId                        Id
}

type clientInfo struct {
	sync.RWMutex
	connections        *list.List
	conntracRecords    *list.List
	videoBufferDepthMs time.Duration
	serial             Id
	eventStream        chan event.Event
	connectionByConnId map[Id]*connection
}

type Id uint64

func (i Id) String() string {
	return fmt.Sprintf("%d", i)
}

/*
 ConnectionTracker is receiving updates from multiple sources:
  - requests to register data streams
  - updates from NAT connection table so that we could see real IPs of clients
  - updates from clients regarding which time ranges they currently cache
  - requests to export specific time ranges for certain clients
*/
type ConnectionTracker struct {
	sync.RWMutex

	qRegisterPfring util.Queue // *cmdRegisterStreamLeg
	qRegisterTls    util.Queue // *cmdRegisterDuplex
	qExport         util.Queue // *cmdExport
	qExportJobs     util.Queue // *jobExport
	qExportComplete util.Queue // *jobExport
	qNatConntrac    util.Queue // *conntrac.ConntracRec

	listenPorts         map[gopacket.Endpoint]bool
	natInterfaceAddress gopacket.Endpoint

	clients         map[gopacket.Endpoint]*clientInfo             // holds info after association
	clients_pending map[gopacket.Endpoint][]*cmdRegisterStreamLeg // holds connection registration events before they're mapped to NAT
	natPortmap      map[gopacket.Endpoint]*conntrac.ConntracRec

	idCounter uint64
}

func (ct *ConnectionTracker) NewId() Id {
	return (Id)(atomic.AddUint64(&ct.idCounter, 1))
}

func (ct *ConnectionTracker) debugDump() {
	if !glog.V(vDumpState) {
		return
	}

	glog.Infof("************** CONNECTION TRACKER ********** ")
	glog.Infof("  Pending ports")
	for port, ctr := range ct.natPortmap {
		if ctr != nil {
			glog.Infof("     %v: %v", port, ctr.String())
		}
	}
	glog.Infof("  Pending connections")
	for ep, legs := range ct.clients_pending {
		if (legs != nil) && (len(legs) > 0) {
			glog.Infof("    %v: %v", ep, legs)
		}
	}

	glog.Infof("******************************************** ")
}

func NewConnectionTracker(listenPorts map[gopacket.Endpoint]bool, natIP net.IP) *ConnectionTracker {
	ct := &ConnectionTracker{
		qRegisterPfring: util.NewMonitoredQueue("PF_RING StreamsQueue"),
		qRegisterTls:    util.NewMonitoredQueue("TLS StreamQueue"),
		qExport:         util.NewMonitoredQueue("ExportCommandsQueue"),
		qExportJobs:     util.NewMonitoredQueue("ExportJobsQueue"),
		qExportComplete: util.NewMonitoredQueue("ExportCompleteQueue"),
		//qExportUploads:  util.NewMonitoredQueue("ExportUploadQueue"),
		qNatConntrac: util.NewMonitoredQueue("NATConnectionTrackerQueue"),

		listenPorts:         listenPorts,
		natInterfaceAddress: layers.NewIPEndpoint(natIP),

		clients:         make(map[gopacket.Endpoint]*clientInfo),
		clients_pending: make(map[gopacket.Endpoint][]*cmdRegisterStreamLeg),
		natPortmap:      make(map[gopacket.Endpoint]*conntrac.ConntracRec),
	}

	conntrac.NewConntrac(ct.qNatConntrac)
	go util.SafeRun(ct.run)

	return ct
}

// check we have record for this client, create one if we don't
func (ct *ConnectionTracker) getClientInfo(client gopacket.Endpoint) *clientInfo {
	// TODO: we somehow need to forcibly cleanup the IP address cache after VPN is down
	//  	 otherwise we might start observing someone else's data

	ct.RLock()

	clientEx, there := ct.clients[client]
	if !there {
		ct.RUnlock()

		ct.Lock()
		defer ct.Unlock()

		clientEx = &clientInfo{
			connections:        new(list.List),
			conntracRecords:    new(list.List),
			videoBufferDepthMs: time.Minute,
			serial:             ct.NewId(),
			connectionByConnId: make(map[Id]*connection)}
		ct.clients[client] = clientEx

		glog.Infof("New client registered %v", client)
		return clientEx
	} else {
		ct.RUnlock()
		return clientEx
	}
}

type cmdRegisterStreamLeg struct {
	net, transport gopacket.Flow
	streamData     *SegmentedStream
	streamType     int
}

type cmdRegisterStreamDuplex struct {
	track             *conntrac.ConntracRec
	request, response *SegmentedStream
	connId            Id
}

/* registers a specific data flow, which later could be matched with other leg and NAT conntrack table in order to
   detect actual origin and destination
*/
func (ct *ConnectionTracker) RegisterStreamLeg(net, transport gopacket.Flow, streamData *SegmentedStream, streamType int) {
	ct.qRegisterPfring.Push(&cmdRegisterStreamLeg{net, transport, streamData, streamType})
}

func (ct *ConnectionTracker) RegisterStreamDuplex(connId Id, request, response *SegmentedStream, track *conntrac.ConntracRec) {
	ct.qRegisterTls.Push(&cmdRegisterStreamDuplex{request: request, response: response, track: track, connId: connId})
}

func (ct *ConnectionTracker) run() {
	selector := util.NewSelector(map[util.Queue]util.Handler{
		ct.qRegisterPfring: ct.doRegisterStreamLeg,
		ct.qRegisterTls:    ct.doRegisterStreamDuplex,
		ct.qExport:         ct.doExportCmd,
		ct.qExportJobs:     ct.doExportJob,
		ct.qExportComplete: ct.doExportComplete,
		ct.qNatConntrac:    ct.doNatConntracUpdate,
	})

	selector.LoopWithTimeout(cTrimDeadline, ct.doTrimming)
}
