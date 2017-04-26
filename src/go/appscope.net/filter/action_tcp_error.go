package filter

import (
    "errors"
    "github.com/golang/glog"
    "io"
    "net/http"
)

const (
    ACTION_TCP_ERROR_POINT_BEFORE_HEADER = "BeforeHeader"
    ACTION_TCP_ERROR_POINT_IN_BODY       = "InBody"
    ACTION_TCP_ERROR_POINT_AFTER_BODY    = "AfterBody"
)

var ACTION_TCP_ERROR_ERR = errors.New("Filter action tcp error")

type ActionTcpErrorPayload struct {
    Point string
}

type ActionTcpError struct {
    Type    string
    Payload ActionTcpErrorPayload
}

type ActionTcpErrorReadCloser struct {
    originalBody    io.ReadCloser
    bodyBreakOffset int64
    bodyReadOffset  int64
    payload         *ActionTcpErrorPayload
}

func NewActionTcpError() *ActionTcpError {
    action := &ActionTcpError{
        Type: ACTION_TCP_ERROR,
    }

    return action
}

func (action *ActionTcpError) GetType() string {
    return action.Type
}

func (action *ActionTcpError) UpdateRequest(request *http.Request) (err error) {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    switch action.Payload.Point {
        case ACTION_TCP_ERROR_POINT_BEFORE_HEADER:
            err = ACTION_TCP_ERROR_ERR
            
        case ACTION_TCP_ERROR_POINT_IN_BODY:
            rc := &ActionTcpErrorReadCloser {
                originalBody: request.Body,
                bodyReadOffset: 0,
                payload: &action.Payload,
            }
            
            rc.bodyBreakOffset = request.ContentLength /2
            if (rc.bodyBreakOffset < 0) {
                rc.bodyBreakOffset = 0
            }
            
            request.Body = rc
            err = nil
            
        case ACTION_TCP_ERROR_POINT_AFTER_BODY:
            rc := &ActionTcpErrorReadCloser {
                originalBody: request.Body,
                bodyReadOffset: 0,
                payload: &action.Payload,
            }
                      
            rc.bodyBreakOffset = request.ContentLength
            if (rc.bodyBreakOffset < 0) {
                rc.bodyBreakOffset = 0
            }
            
            request.Body = rc
            err = nil
    }
    glog.V(GLOG_LEVEL_ALL).Infof("update request done, err '%v'", err)
    
    return err
}

func (action *ActionTcpError) UpdateResponse(response *http.Response) (err error) {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    switch action.Payload.Point {
        case ACTION_TCP_ERROR_POINT_BEFORE_HEADER:
            err = ACTION_TCP_ERROR_ERR
            
        case ACTION_TCP_ERROR_POINT_IN_BODY:
            rc := &ActionTcpErrorReadCloser {
                originalBody: response.Body,
                bodyReadOffset: 0,
                payload: &action.Payload,
            }
            
            rc.bodyBreakOffset = response.ContentLength /2
            if (rc.bodyBreakOffset < 0) {
                rc.bodyBreakOffset = 0
            }
            
            response.Body = rc
            err = nil
            
        case ACTION_TCP_ERROR_POINT_AFTER_BODY:
            rc := &ActionTcpErrorReadCloser {
                originalBody: response.Body,
                bodyReadOffset: 0,
                payload: &action.Payload,
            }
            
            rc.bodyBreakOffset = response.ContentLength
            if (rc.bodyBreakOffset < 0) {
                rc.bodyBreakOffset = 0
            }
            
            response.Body = rc
            err = nil
    }
    glog.V(GLOG_LEVEL_ALL).Infof("update response done, err '%v'", err)
    
    return err
}

func (rc *ActionTcpErrorReadCloser) Read(p []byte) (n int, err error) {
    glog.V(GLOG_LEVEL_ALL).Infof("read, requested %v bytes", len(p))
    
    if (rc.bodyReadOffset < rc.bodyBreakOffset) {
        n, err = rc.originalBody.Read(p)
        glog.V(GLOG_LEVEL_ALL).Infof("read, got %v bytes, err % v", n, err)
        
        if ((err == nil) || (err == io.EOF)) {
            if ((rc.bodyReadOffset + int64(n)) > rc.bodyBreakOffset) {
                glog.V(GLOG_LEVEL_ALL).Infof("read, reached break offset %v , next read generates error", rc.bodyBreakOffset)
                n = int(rc.bodyBreakOffset - rc.bodyReadOffset)
                rc.bodyReadOffset = rc.bodyBreakOffset
            } else {
                rc.bodyReadOffset += int64(n)
            }
        }
    } else {
        glog.V(GLOG_LEVEL_ALL).Infof("breaking the connection, total read so far %v, break offset %v", rc.bodyReadOffset, rc.bodyBreakOffset)
        n = 0
        err = ACTION_TCP_ERROR_ERR
    }
    glog.V(GLOG_LEVEL_ALL).Infof("read done, n %v, err '%v'", n, err)
    
    return n, err
}

func (rc *ActionTcpErrorReadCloser) Close() error {
    glog.V(GLOG_LEVEL_ALL).Infof("close")
    
    if (rc.payload.Point == ACTION_TCP_ERROR_POINT_AFTER_BODY) {
        return ACTION_TCP_ERROR_ERR
    } else {
        return nil
    }
}
