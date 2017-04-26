package filter

import (
    "github.com/golang/glog"
    "net/http"
)

type ActionReplaceHeaderPayload struct {
    Merge    bool
    Fields []KeyValue
}

type ActionReplaceHeader struct {
    Type    string
    Payload ActionReplaceHeaderPayload
}

func NewActionReplaceHeader(header []byte) *ActionReplaceHeader {
    action := &ActionReplaceHeader{
        Type: ACTION_REPLACE_HEADER,
    }

    return action
}

func (action *ActionReplaceHeader) GetType() string {
    return action.Type
}

func (action *ActionReplaceHeader) UpdateRequest(request *http.Request) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    if (!action.Payload.Merge) {
        if (len(action.Payload.Fields) != 0) {
            request.Header = make(http.Header)
        } else {
            request.Header = nil
        }
    }
    
    for i := 0; i < len(action.Payload.Fields); i++ {
        request.Header.Set(action.Payload.Fields[i].Key, action.Payload.Fields[i].Value)
    }
    
    return nil
}

func (action *ActionReplaceHeader) UpdateResponse(response *http.Response) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    if (!action.Payload.Merge) {
        if (len(action.Payload.Fields) != 0) {
            response.Header = make(http.Header)
        } else {
            response.Header = nil
        }
    }
    
    for i := 0; i < len(action.Payload.Fields); i++ {
        response.Header.Set(action.Payload.Fields[i].Key, action.Payload.Fields[i].Value)
    }
    
    return nil
}
