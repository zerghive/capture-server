package filter

import (
    "bytes"
    "github.com/golang/glog"
    "io"
    "net/http"
    "sync"
)

const (
    GLOG_LEVEL_ALL = 100
)

type FilterGuid2CompoundFilterMap map[string] CompoundFilter

type Filter struct {
    Triggers    []Trigger
    Actions     []Action
    MaxCount    int
    valueCounts map[string] int
    mutex       sync.RWMutex
}

type CompoundFilter struct {
    RequestFilter  *Filter
    ResponseFilter *Filter
}

func NewFilter(triggers []Trigger, actions []Action, maxCount int) *Filter {
    filter := &Filter{
        Triggers  : triggers,
        Actions   : actions,
        MaxCount  : maxCount,
    }
    
    if (maxCount != 0) {
        filter.valueCounts = make(map[string] int)
    }

    return filter
}

func NewCompoundFilter(request *Filter, response *Filter) *CompoundFilter {
    filter := &CompoundFilter{
        RequestFilter  : request,
        ResponseFilter : response,
    }

    return filter
}

func (filter *CompoundFilter) GetSyntheticResponse() (response *http.Response) {
    for i := 0; i < len(filter.RequestFilter.Actions); i++ {
        if (IsSynthetic(filter.RequestFilter.Actions[i])) {
            if (response == nil) {
                response = new(http.Response)
            }
            
            filter.RequestFilter.Actions[i].UpdateResponse(response)
        }
    }

    return response
}

func (filter *Filter) TryRequest(request *http.Request) (triggered bool, triggeredValues []TriggerInfo, count int) {
    var regexpedValue, originalValue string
    var countingValue bytes.Buffer
    
    triggered = true
    
    // logical AND for triggers
    for i := 0; (i < len(filter.Triggers)) && triggered; i++ {
        triggered, originalValue, regexpedValue = filter.Triggers[i].TryRequest(request)

        if (triggered) {
            triggeredValues = append(triggeredValues, 
                TriggerInfo {
                    Key          : filter.Triggers[i].Key, 
                    OriginalValue: originalValue,
                    RegexpedValue: regexpedValue,
                    Counting     : filter.Triggers[i].Counting,
                })
        }
        
        if (filter.Triggers[i].Counting && (filter.MaxCount != 0)) {
            // creating complex string of triggered values
            if (countingValue.Len() != 0) {
                countingValue.WriteString(",")
            }
            countingValue.WriteString(regexpedValue)
        }
    }
    
    if (triggered && (filter.MaxCount != 0)) {
        filter.mutex.Lock()

        if tmpCount, found := filter.valueCounts[countingValue.String()]; found {
            filter.valueCounts[countingValue.String()] = tmpCount +1
        } else {
            filter.valueCounts[countingValue.String()] = 1
        }

        count = filter.valueCounts[countingValue.String()]
        
        triggered = (count <= filter.MaxCount)

        filter.mutex.Unlock()
        
        glog.V(GLOG_LEVEL_ALL).Infof("REQUEST key '%v', count %v, max count %v", countingValue.String(), count, filter.MaxCount)
    }
        
    return triggered, triggeredValues, count
}

func (filter *Filter) TryResponse(response *http.Response) (triggered bool, triggeredValues []TriggerInfo, count int) {
    var regexpedValue, originalValue string
    var countingValue bytes.Buffer
    
    triggered = true
    
    // logical AND for triggers
    for i := 0; (i < len(filter.Triggers)) && triggered; i++ {
        triggered, originalValue, regexpedValue = filter.Triggers[i].TryResponse(response)

        if (triggered) {
            triggeredValues = append(triggeredValues, 
                TriggerInfo {
                    Key          : filter.Triggers[i].Key, 
                    OriginalValue: originalValue,
                    RegexpedValue: regexpedValue,
                    Counting     : filter.Triggers[i].Counting,
                })
        }
        
        if (filter.Triggers[i].Counting && (filter.MaxCount != 0)) {
            // creating complex string of triggered values
            if (countingValue.Len() != 0) {
                countingValue.WriteString(",")
            }
            countingValue.WriteString(regexpedValue)
        }
    }
    
    if (triggered && (filter.MaxCount != 0)) {
        filter.mutex.Lock()

        if tmpCount, found := filter.valueCounts[countingValue.String()]; found {
            filter.valueCounts[countingValue.String()] = tmpCount +1
        } else {
            filter.valueCounts[countingValue.String()] = 1
        }

        count = filter.valueCounts[countingValue.String()]
        
        triggered = (count <= filter.MaxCount)

        filter.mutex.Unlock()
        
        glog.V(GLOG_LEVEL_ALL).Infof("RESPONSE key '%v', count %v, max count %v", countingValue.String(), count, filter.MaxCount)
    }
    
    return triggered, triggeredValues, count
}

func (filter *Filter) WriteRequest(request *http.Request, writer io.Writer) (err error) {
    err = nil
       
    for i := 0; (i < len(filter.Actions)) && (err == nil); i++ {
        err = filter.Actions[i].UpdateRequest(request)
    }
    
    if (err == nil) {
        err = WriteRequest(request, writer)
    } else {
        request.Body.Close()
    }

    return err
}

func (filter *Filter) WriteResponse(response *http.Response, writer io.Writer) (err error) {
    err = nil
     
    for i := 0; (i < len(filter.Actions)) && (err == nil); i++ {
        err = filter.Actions[i].UpdateResponse(response)
    }
    
    if (err == nil) {
        err = WriteResponse(response, writer)
    } else {
        response.Body.Close()
    }
    
    return err
}
