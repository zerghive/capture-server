package filter

import (
    "github.com/golang/glog"
    "io"
    "net/http"
)

const (
    ACTION_SYNTHETIC_RESPONSE_PROTOMAJOR = 1
    ACTION_SYNTHETIC_RESPONSE_PROTOMINOR = 1
)

type ActionSyntheticResponsePayload struct {
    DataType string
    Data     string
    Fields   []KeyValue
    Status   int
}

type ActionSyntheticResponse struct {
    Type           string
    Payload        ActionSyntheticResponsePayload
}

type ActionSyntheticResponseReadCloser struct {
    bodyReadOffset int
    payload        *ActionSyntheticResponsePayload
}

func NewActionSyntheticResponse() *ActionSyntheticResponse {    
    action := &ActionSyntheticResponse{
        Type: ACTION_SYNTHETIC_RESPONSE,
    }

    return action
}

func (action *ActionSyntheticResponse) GetType() string {
    return action.Type
}

func (action *ActionSyntheticResponse) UpdateRequest(request *http.Request) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    return nil
}

func (action *ActionSyntheticResponse) UpdateResponse(response *http.Response) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    rc := &ActionSyntheticResponseReadCloser {
        bodyReadOffset: 0,
        payload       : &action.Payload,
    }
    
    response.ProtoMajor    = ACTION_SYNTHETIC_RESPONSE_PROTOMAJOR
    response.ProtoMinor    = ACTION_SYNTHETIC_RESPONSE_PROTOMINOR
    response.StatusCode    = action.Payload.Status
    response.Status        = StatusStrings[action.Payload.Status]
    response.ContentLength = int64(len(action.Payload.Data))
    
    if (len(action.Payload.Fields) != 0) {
        response.Header = make(http.Header)
    } else {
        response.Header = nil
    }
    
    for i := 0; i < len(action.Payload.Fields); i++ {
        response.Header.Set(action.Payload.Fields[i].Key, action.Payload.Fields[i].Value)
    }

    response.Body = rc
    
    return nil
}

func (rc *ActionSyntheticResponseReadCloser) Read(p []byte) (n int, err error) {
    glog.V(GLOG_LEVEL_ALL).Infof("Read, requested %v bytes", len(p))
    
    var copyBytes = 0
    err = nil
    
    if ((len(rc.payload.Data) - rc.bodyReadOffset) < len(p)) {
        copyBytes = len(rc.payload.Data) - rc.bodyReadOffset
    } else {
        copyBytes = len(p)
    }
    
    if (copyBytes != 0) {
        copy(p, []byte(rc.payload.Data[rc.bodyReadOffset: (rc.bodyReadOffset + copyBytes)]))
        glog.V(GLOG_LEVEL_ALL).Infof("Read, copied data: %v", []byte(rc.payload.Data[rc.bodyReadOffset: (rc.bodyReadOffset + copyBytes)]))
        
        rc.bodyReadOffset += copyBytes
    } else {
        err = io.EOF
    }
    
    glog.V(GLOG_LEVEL_ALL).Infof("Read, copied %v bytes, err %v", copyBytes, err)
    
    return copyBytes, err
}

func (rc *ActionSyntheticResponseReadCloser) Close() error {
    glog.V(GLOG_LEVEL_ALL).Infof("Close")
    return nil
}
