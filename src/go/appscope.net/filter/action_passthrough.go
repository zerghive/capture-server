package filter

import (
    "github.com/golang/glog"
    "net/http"
)

type ActionPassthroughPayload struct {
}

type ActionPassthrough struct {
    Type    string
    Payload ActionPassthroughPayload
}

func NewActionPassthrough() *ActionPassthrough {
    action := &ActionPassthrough{
        Type: ACTION_PASSTHROUGH,
    }

    return action
}

func (action *ActionPassthrough) GetType() string {
    return action.Type
}

func (action *ActionPassthrough) UpdateRequest(request *http.Request) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)
    
    return nil
}

func (action *ActionPassthrough) UpdateResponse(response *http.Response) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v", action.Type)

    return nil
}
