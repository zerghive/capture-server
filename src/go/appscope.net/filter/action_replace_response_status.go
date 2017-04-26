package filter

import (
    "github.com/golang/glog"
    "net/http"
)

type ActionReplaceResponseStatusPayload struct {
    Status int
}

type ActionReplaceResponseStatus struct {
    Type    string
    Payload ActionReplaceResponseStatusPayload
}

func NewActionReplaceResponseStatus() *ActionReplaceResponseStatus {    
    action := &ActionReplaceResponseStatus{
        Type: ACTION_REPLACE_RESPONSE_STATUS,
    }

    return action
}

func (action *ActionReplaceResponseStatus) GetType() string {
    return action.Type
}

func (action *ActionReplaceResponseStatus) UpdateRequest(request *http.Request) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)

    return nil
}

func (action *ActionReplaceResponseStatus) UpdateResponse(response *http.Response) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    response.StatusCode = action.Payload.Status
    response.Status =  StatusStrings[action.Payload.Status]
    
    return nil
}
