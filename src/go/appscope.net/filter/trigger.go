package filter

import (
    "github.com/golang/glog"
    "net/http"
    "regexp"
    "strconv"
)

const (
    TRIGGER_TYPE_REQUEST    = "Request"
    TRIGGER_TYPE_RESPONSE   = "Response"
    TRIGGER_TYPE_HEADER     = "Header"
)

type Trigger struct {
    Type       string
    Key        string
    Regexp     string
    Counting   bool
    regexp    *regexp.Regexp
    raised     bool
}

func NewTrigger(typ string, key string, regex string, counting bool) *Trigger {
    trigger := &Trigger{
        Type     : typ,
        Key      : key,
        Regexp   : regex,
        Counting : counting,
    }
    
    regexp, err := regexp.Compile(regex)
    if (err == nil) {
        trigger.regexp = regexp
    }

    trigger.raised = (len(typ) == 0) && (len(key) == 0) && (len(regex) == 0)

    return trigger
}

func NewRaisedTrigger() *Trigger {
    trigger := &Trigger{
        raised   : true,
    }
    
    return trigger
}

func (trigger *Trigger) TryRequest(request *http.Request) (triggered bool, requestValue string, regexpedValue string) {
    glog.V(GLOG_LEVEL_ALL).Infof("TRY REQUEST trigger key = '%s', regexp string = '%s', regexp '%v'", trigger.Key, trigger.Regexp, trigger.regexp)
    
    triggered = false
    
    if (!trigger.raised) {        
        if (trigger.regexp != nil) {
            switch trigger.Type {
                case TRIGGER_TYPE_REQUEST:
                    switch trigger.Key {
                        case "Method"    : requestValue = request.Method
                        case "Host"      : requestValue = request.Host
                        case "RemoteAddr": requestValue = request.RemoteAddr
                        case "RequestURI": requestValue = request.RequestURI
                    }

                case TRIGGER_TYPE_HEADER:
                    requestValue = request.Header.Get(trigger.Key)
            }
            
            triggered = trigger.regexp.MatchString(requestValue)

            regexpedValue = trigger.regexp.FindString(requestValue)
            if (len(regexpedValue) == 0) {
                // when regexp string is "", consider the regexped value to be request value as is
                regexpedValue = requestValue
            }
        }
    }
    
    glog.V(GLOG_LEVEL_ALL).Infof("counting trigger = '%v', raised = '%v', triggered = '%v'", trigger.Counting, trigger.raised, triggered)
    
    triggered = triggered || trigger.raised
    
    return triggered, requestValue, regexpedValue
}

func (trigger *Trigger) TryResponse(response *http.Response) (triggered bool, responseValue string, regexpedValue string) {
    glog.V(GLOG_LEVEL_ALL).Infof("TRY RESPONSE trigger key = '%s', regexp string = '%s', regexp '%v'", trigger.Key, trigger.Regexp, trigger.regexp)
    
    triggered = false
    
    if (!trigger.raised) {
        if (trigger.regexp != nil) {
            switch trigger.Type {
                case TRIGGER_TYPE_RESPONSE:
                    switch trigger.Key {
                        case "Status"    : responseValue = response.Status
                        case "StatusCode": responseValue = strconv.Itoa(response.StatusCode)
                        case "Proto"     : responseValue = response.Proto
                        case "ProtoMajor": responseValue = strconv.Itoa(response.ProtoMajor)
                        case "ProtoMinor": responseValue = strconv.Itoa(response.ProtoMinor)
                    }

                case TRIGGER_TYPE_HEADER:
                    responseValue = response.Header.Get(trigger.Key)
            }
            
            triggered = trigger.regexp.MatchString(responseValue)
            
            regexpedValue = trigger.regexp.FindString(responseValue)
            if (len(regexpedValue) == 0) {
                // when regexp string is "", consider the regexped value to be request value as is
                regexpedValue = responseValue
            }
        }
    }

    glog.V(GLOG_LEVEL_ALL).Infof("counting trigger = '%v', raised = '%v', triggered = '%v'", trigger.Counting, trigger.raised, triggered)

    triggered = triggered || trigger.raised
    
    return triggered, responseValue, regexpedValue
}
