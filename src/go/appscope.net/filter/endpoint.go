package filter

import (
    "github.com/golang/glog"
    "io"
    "net"
    "net/http"
)

type SequenceMarker interface {
    MarkStart()
    MarkEnd()
}

type EndpointControlMessage struct {
    requestPassthrough bool
    response           *http.Response
    responseEndpoint   io.Writer
}

type Endpoint struct {
     cache          []byte
     connection     net.Conn
     controlChannel chan EndpointControlMessage
     currentMessage EndpointControlMessage
     startMarked    bool
}

func NewEndpoint(conn net.Conn) *Endpoint {
    endpoint := &Endpoint{
        connection : conn,
        startMarked: false,
    }
    
    endpoint.controlChannel = make(chan EndpointControlMessage, 1)

    return endpoint
}

func (endpoint *Endpoint) MarkStart() {
    glog.V(GLOG_LEVEL_ALL).Infof("mark start")
       
    endpoint.currentMessage, _ = <- endpoint.controlChannel
    
    glog.V(GLOG_LEVEL_ALL).Infof("sending cache %v, data size %d", endpoint.currentMessage.requestPassthrough, len(endpoint.cache))
    if (endpoint.currentMessage.requestPassthrough) {
        nLeft := len(endpoint.cache)
        
        for (nLeft > 0) {    
            consumed, err := endpoint.connection.Write(endpoint.cache)
            
            if (err == nil) {
                nLeft -= consumed
            } else {
                glog.V(GLOG_LEVEL_ALL).Infof("writing cache, left bytes %d, ERROR %v", nLeft, err)
                break;
            }
        }
    }
    
    endpoint.cache = nil
    endpoint.startMarked = true
}

func (endpoint *Endpoint) MarkEnd() {
    glog.V(GLOG_LEVEL_ALL).Infof("mark end")
    if (!endpoint.currentMessage.requestPassthrough && endpoint.startMarked) {
        if (endpoint.currentMessage.response != nil) {
            err := endpoint.currentMessage.response.Write(endpoint.currentMessage.responseEndpoint)
            glog.V(GLOG_LEVEL_ALL).Infof("writing synthetic response, err = %v", err)
        }
    }
    
    endpoint.startMarked = false
}

func (endpoint *Endpoint) Control(passthrough bool, response *http.Response, writer io.Writer) {
    glog.V(GLOG_LEVEL_ALL).Infof("control passthrough %v, resp %v", passthrough, response)
    endpoint.controlChannel <- EndpointControlMessage{passthrough, response, writer}
}

func (endpoint *Endpoint) Write(bytes []byte) (n int, err error) {
    if (endpoint.startMarked) {
        if (endpoint.currentMessage.requestPassthrough) {
            glog.V(GLOG_LEVEL_ALL).Infof("passthrough bytes to endpoint %v", len(bytes))
            return endpoint.connection.Write(bytes)
        } else {
            glog.V(GLOG_LEVEL_ALL).Infof("drop bytes %v", len(bytes))
            return len(bytes), nil
        }
    } else {
        endpoint.cache = append(endpoint.cache, bytes...)
        glog.V(GLOG_LEVEL_ALL).Infof("caching data size %d, total cached %d", len(bytes), len(endpoint.cache))
        return len(bytes), nil
    }
}
