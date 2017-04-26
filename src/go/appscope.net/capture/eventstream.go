package capture

import (
	"github.com/golang/glog"
	"github.com/google/gopacket"

	"appscope.net/conntrac"
	"appscope.net/event"
	"appscope.net/util"

	"time"
)

var cClientEventTimeout time.Duration = time.Second

const (
	cEventStreamBuffer = 100

	HttpConnectionStateRecording = "recording"
)

/*
 * used to post live update events to connected web interface
 */

type TcpConnEvent struct {
	Id              int
	Src, Dst, State string
}

type HttpConnectionEvent struct {
	Id          Id `json:",string"`
	Host, State string
	TcpConnId   string
}

type ConnectionErrorEvent struct {
	Id         Id `json:",string"`
	Error      conntrac.SegmentError
	Tls        bool
	ServerName string
	ServerAddr string
}

type HttpRequestEvent struct {
	Id, HttpConnId Id `json:",string"`
	RequestId      int
	Info           *RequestInfo
	Error          error
}

type HttpResponseEvent struct {
	Id, HttpConnId Id  `json:",string"`
	RequestId      int // within connection
	Info           *ResponseInfo
	Error          error
} 

func (t *TcpConnEvent)         Type() string { return "tcp" }
func (t *HttpConnectionEvent)  Type() string { return "http" }
func (t *ConnectionErrorEvent) Type() string { return "conn_error" }
func (t *HttpRequestEvent)     Type() string { return "http_request" }
func (t *HttpResponseEvent)    Type() string { return "http_response" }

func (ct *ConnectionTracker) PostClientEvent(client gopacket.Endpoint, eventSource func() event.Event) {
	stream := ct.getEventStream(client)
	if stream == nil {
		return
	}

	util.SafeRun(func() {
		event := eventSource()
		select {
		case stream <- event:
			return
		case <-time.After(cClientEventTimeout):
			glog.Errorf("Event to %v discarded due to timeout", client)
		}
	})
}

func (ct *ConnectionTracker) getEventStream(client gopacket.Endpoint) chan event.Event {
	clientInfo := ct.getClientInfo(client)

	clientInfo.RLock()
	defer clientInfo.RUnlock()

	return clientInfo.eventStream
}

func (ct *ConnectionTracker) installEventStream(client gopacket.Endpoint) chan event.Event {
	clientInfo := ct.getClientInfo(client)

	clientInfo.Lock()
	defer clientInfo.Unlock()

	if clientInfo.eventStream != nil {
		glog.Infof("Replacing existing event stream for %v", client)
		close(clientInfo.eventStream)
		clientInfo.eventStream = nil
	}

	clientInfo.eventStream = make(chan event.Event, cEventStreamBuffer)
	return clientInfo.eventStream
}

func (ct *ConnectionTracker) shutdownEventStream(client gopacket.Endpoint) {
	clientInfo := ct.clients[client]

	clientInfo.Lock()
	defer clientInfo.Unlock()

	if clientInfo.eventStream != nil {
		close(clientInfo.eventStream)
		clientInfo.eventStream = nil
	}
}
