package filter

import (
    "errors"
    "github.com/golang/glog"
    "net/http"
    "time"
)

const (
    ACTION_DELAY_DESCRIPTION = "Filter action infinite delay"
)

type ActionDelayPayload struct {
    DelayMs int
}

type ActionDelay struct {
    Type    string
    Payload ActionDelayPayload
}

func NewActionDelay(ms int) *ActionDelay {
    action := &ActionDelay{
        Type : ACTION_DELAY,
    }
    
    action.Payload.DelayMs = ms

    return action
}

func (action *ActionDelay) GetType() string {
    return action.Type
}

func (action *ActionDelay) UpdateRequest(request *http.Request) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v %v ms", action.Type, action.Payload.DelayMs)

    if (action.Payload.DelayMs == 0) {
        // special value "infinite", just drop it
        return errors.New(ACTION_DELAY_DESCRIPTION)
    } else {
        time.Sleep(time.Duration(action.Payload.DelayMs) * time.Millisecond)
        
        return nil
    }
}

func (action *ActionDelay) UpdateResponse(response *http.Response) error {
    glog.V(GLOG_LEVEL_ALL).Infof("%v %v ms", action.Type, action.Payload.DelayMs)

    if (action.Payload.DelayMs == 0) {
        // special value "infinite", just drop it
        return errors.New(ACTION_DELAY_DESCRIPTION)
    } else {
        time.Sleep(time.Duration(action.Payload.DelayMs) * time.Millisecond)
        
        return nil
    }
}
