package filter

import (
    "appscope.net/event"
    "bufio"
    "github.com/golang/glog"
    "github.com/google/gopacket"
    "io"
    "net"
    "net/http"
    "sync"
)

type ChannelMessage struct {
    filter *CompoundFilter
    guid string
}

type Modifier struct {
    clientConnection  net.Conn
    serverConnection  net.Conn
    
    requestPipeReader *io.PipeReader
    requestPipeWriter *io.PipeWriter
    
    responsePipeReader *io.PipeReader
    responsePipeWriter *io.PipeWriter
    
    modifiedRequestReader  io.Reader
    modifiedResponseReader io.Reader
    
    serverEndpoint *Endpoint
    
    attachedFilters    FilterGuid2CompoundFilterMap
    passthroughFilter  CompoundFilter
    filterChannel      chan ChannelMessage
    
    eventStream        event.EventStream
    
    requestIdChannel   chan int
    responseIdChannel  chan int
    
    clientId           string
    clientTcpAddr      gopacket.Endpoint
    mutex              sync.RWMutex
}

func NewModifier(lId string, lTcpAddr gopacket.Endpoint, lConn, rConn net.Conn, stream event.EventStream) *Modifier {
    glog.V(GLOG_LEVEL_ALL).Infof("new MODIFIER, client Id: %v", lId)
    
    modifier := &Modifier{
        clientConnection: lConn,
        serverConnection: rConn,
        clientId        : lId,
        clientTcpAddr   : lTcpAddr,
        eventStream     : stream,
    }
    
    modifier.modifiedRequestReader  = NewReader(modifier.ReadModifiedRequest)
    modifier.modifiedResponseReader = NewReader(modifier.ReadModifiedResponse)
    
    modifier.serverEndpoint = NewEndpoint(rConn)
   
    modifier.filterChannel      = make(chan ChannelMessage, 1)
    
    modifier.requestIdChannel   = make(chan int, 1)
    modifier.responseIdChannel  = make(chan int, 1)
    
    modifier.requestPipeReader, modifier.requestPipeWriter = io.Pipe()
    modifier.responsePipeReader, modifier.responsePipeWriter = io.Pipe()
    
    // need passthrough filter as default when no other filters are triggered    
    modifier.passthroughFilter = NewPassthroughFilter()

    return modifier
}

func NewPassthroughFilter() CompoundFilter {    
    return *NewCompoundFilter(
        NewFilter([]Trigger{*NewRaisedTrigger()}, []Action{NewActionPassthrough()}, 0), 
        NewFilter([]Trigger{*NewRaisedTrigger()}, []Action{NewActionPassthrough()}, 0))
}

func (modifier *Modifier) SetFilters(filters FilterGuid2CompoundFilterMap) error {
    glog.V(GLOG_LEVEL_ALL).Infof("set filters, client IP: %v, filters %v", modifier.clientId, filters)
    modifier.mutex.Lock()
    modifier.attachedFilters = filters
    modifier.mutex.Unlock()
    
    return nil
}

func (modifier *Modifier) SetRequestId(id int) {
    modifier.requestIdChannel <- id
}

func (modifier *Modifier) SetResponseId(id int) {
    modifier.responseIdChannel <- id
}

func (modifier *Modifier) GetClientId() string {
    return modifier.clientId
}

func (modifier *Modifier) GetTriggeredFilterByRequest(request *http.Request) (info FilterInfo, filter *CompoundFilter)  {   
    modifier.mutex.RLock()
    for filterGuid, currentFilter := range modifier.attachedFilters {
        if triggered, triggeredValues, currentCount := currentFilter.RequestFilter.TryRequest(request); triggered  {
            filter = &currentFilter
            
            info = FilterInfo {
                Guid           : filterGuid,
                Count          : currentCount,
                TriggeredValues: triggeredValues,
            }
            break
        }
    }
    modifier.mutex.RUnlock()
    
    return info, filter
}

func (modifier *Modifier) RequestSrc() io.Reader {
    return modifier.modifiedRequestReader
}

func (modifier *Modifier) RequestDst() *Endpoint {
    return modifier.serverEndpoint
}

func (modifier *Modifier) ResponseSrc() io.Reader {
    return modifier.modifiedResponseReader
}

func (modifier *Modifier) ResponseDst() io.Writer {
    return modifier.clientConnection
}

func (modifier *Modifier) ReadModifiedRequest(bytes []byte) (n int, err error) {
    return modifier.requestPipeReader.Read(bytes)
}

func (modifier *Modifier) ReadModifiedResponse(bytes []byte) (n int, err error) {
    return modifier.responsePipeReader.Read(bytes)
}

func (modifier *Modifier) WriteRequest(request *http.Request) (event *HttpRequestEvent, err error) {
    glog.V(GLOG_LEVEL_ALL).Infof("write REQUEST, host %s, uri %s", request.Host, request.RequestURI)
        
    info, filter := modifier.GetTriggeredFilterByRequest(request)
    
    if (filter != nil) {
        event = &HttpRequestEvent {
            Info: info,
        }
    } else {
        filter = &modifier.passthroughFilter
    }
    
    syntheticResponse := filter.GetSyntheticResponse()
    if (syntheticResponse == nil) {
        // filter only real responses from server
        modifier.filterChannel <- ChannelMessage{filter, info.Guid}
    }
    glog.V(GLOG_LEVEL_ALL).Infof("internal request write, synthetic response = %v", syntheticResponse)
    
    modifier.serverEndpoint.Control(syntheticResponse == nil, syntheticResponse, modifier.responsePipeWriter)
    
    err = filter.RequestFilter.WriteRequest(request, modifier.requestPipeWriter)
    if (err == nil) {
        id := <- modifier.requestIdChannel

        if (event != nil) {
            event.RequestId = id
        }
    }
    
    return event, err
}

func (modifier *Modifier) WriteResponse(response *http.Response) (event *HttpResponseEvent, err error) {
    glog.V(GLOG_LEVEL_ALL).Infof("write RESPONSE, header %s", response.Header)
    
    message := <- modifier.filterChannel
    
    if triggered, triggeredValues, currentCount := message.filter.ResponseFilter.TryResponse(response); triggered {
        if (message.filter.ResponseFilter != modifier.passthroughFilter.ResponseFilter) {
            event = &HttpResponseEvent {
                Info: FilterInfo {
                    Guid           : message.guid,
                    Count          : currentCount,
                    TriggeredValues: triggeredValues,
                },
            }
        }
    }
    
    err = message.filter.ResponseFilter.WriteResponse(response, modifier.responsePipeWriter)
    
    if (err == nil) {
        id := <- modifier.responseIdChannel
    
        if (event != nil) {
            event.RequestId = id
        }
    }
    
    return event, err
}

func (modifier *Modifier) RunRequestCycle(connId uint64) {
    requestReader := bufio.NewReader(modifier.clientConnection)
    
    for {
        var requestEvent *HttpRequestEvent

        request, err := http.ReadRequest(requestReader)
        if err != nil {
            glog.Warningf("error read request: %v, exiting request cycle", err)
            err = modifier.requestPipeWriter.CloseWithError(err)
            return
        }
        
        requestEvent, err = modifier.WriteRequest(request) 
        if (err == ACTION_TCP_ERROR_ERR) {
            glog.Warningf("error write request: %v, exiting request cycle", err)
            err = modifier.requestPipeWriter.CloseWithError(err)
            return
        }
        
        if (requestEvent != nil) {
            requestEvent.HttpConnId = connId
        
            glog.V(GLOG_LEVEL_ALL).Infof("post request filter event: %v", *requestEvent)
            modifier.eventStream.PostClientEvent(modifier.clientTcpAddr, func() event.Event {
                return requestEvent
            })
        }
    }
}

func (modifier *Modifier) RunResponseCycle(connId uint64) {
    responseReader := bufio.NewReader(modifier.serverConnection)
    
    for {
        var responseEvent *HttpResponseEvent

        response, err := http.ReadResponse(responseReader, nil)	
        if err != nil {
            glog.Warningf("error read response: %v, exiting response cycle", err)
            err = modifier.responsePipeWriter.CloseWithError(err)
            return
        }
        
        responseEvent, err = modifier.WriteResponse(response)
        if (err == ACTION_TCP_ERROR_ERR) {
            glog.Warningf("error write response: %v, exiting response cycle", err)
            err = modifier.responsePipeWriter.CloseWithError(err)
            return
        }
        
        if (responseEvent != nil) {
            responseEvent.HttpConnId = connId
            
            glog.V(GLOG_LEVEL_ALL).Infof("post response filter event: %v", *responseEvent)
            modifier.eventStream.PostClientEvent(modifier.clientTcpAddr, func() event.Event {
                return responseEvent
            })
        }
    }
}
