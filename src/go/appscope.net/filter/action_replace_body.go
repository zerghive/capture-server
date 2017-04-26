package filter

import (
    "github.com/golang/glog"
    "io"
    "net/http"
)

var DATA_CHUNK_SIZE = 64*1024

type ActionReplaceBodyPayload struct {
    DataType string
    Data     string
}

type ActionReplaceBody struct {
    Type    string
    Payload ActionReplaceBodyPayload
}

type ActionReplaceBodyReadCloser struct {
    bodyReadOffset   int
    originalBody     io.ReadCloser
    originalBodyErr  error
    originalBytesBuf []byte
    payload          *ActionReplaceBodyPayload
}

func NewActionReplaceBody() *ActionReplaceBody {    
    action := &ActionReplaceBody{
        Type      : ACTION_REPLACE_BODY,
    }
    
    return action
}

func (action *ActionReplaceBody) GetType() string {
    return action.Type
}

func (action *ActionReplaceBody) UpdateRequest(request *http.Request) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    rc := &ActionReplaceBodyReadCloser{
        bodyReadOffset  : 0,
        originalBody    : request.Body,
        originalBodyErr : nil,
	originalBytesBuf: make([]byte, DATA_CHUNK_SIZE),
        payload         : &action.Payload,
    }
    
    request.ContentLength = int64(len(action.Payload.Data))
    request.Body = rc
    
    return nil
}

func (action *ActionReplaceBody) UpdateResponse(response *http.Response) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    rc := &ActionReplaceBodyReadCloser{
        bodyReadOffset  : 0,
        originalBody    : response.Body,
        originalBodyErr : nil,
	originalBytesBuf: make([]byte, DATA_CHUNK_SIZE),
        payload         : &action.Payload,
    }
    
    response.ContentLength = int64(len(action.Payload.Data))
    response.Body = rc
    
    return nil
}

func (rc *ActionReplaceBodyReadCloser) Read(p []byte) (n int, err error) {
    glog.V(GLOG_LEVEL_ALL).Infof("Read, requested %v bytes", len(p))
    
    var copyBytes = 0
    err = nil
    
    // Read original body if any
    if (rc.originalBodyErr == nil) {
        for {
            _, rc.originalBodyErr = rc.originalBody.Read(rc.originalBytesBuf)
            if (rc.originalBodyErr != nil) {
                glog.V(GLOG_LEVEL_ALL).Infof("Finished reading original body, err %v", err)
                break
            }
        }
    }
    
    // Write configured body
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

func (rc *ActionReplaceBodyReadCloser) Close() error {
    glog.V(GLOG_LEVEL_ALL).Infof("Close")
    
    _ = rc.originalBody.Close()
    
    return nil
}
