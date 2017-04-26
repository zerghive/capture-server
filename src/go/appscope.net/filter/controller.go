package filter

import (
    "appscope.net/event"
    "github.com/golang/glog"
    "github.com/google/gopacket"
    "net"
    "strings"
    "sync"
)

type Controller struct {
    modifiers []*Modifier
    attachedFilters map[string] FilterGuid2CompoundFilterMap
    accessLock sync.Mutex
}

func NewController() *Controller {
    controller := &Controller{
    }
    
    controller.attachedFilters = make(map[string] FilterGuid2CompoundFilterMap)

    return controller
}

func (controller *Controller) CreateModifier(clientId string, clientTcpAddr gopacket.Endpoint, lConn, rConn net.Conn, stream event.EventStream) *Modifier {
    modifier := NewModifier(clientId, clientTcpAddr, lConn, rConn, stream)
    
    controller.accessLock.Lock()
    
    filters := controller.attachedFilters[clientId]
    glog.V(GLOG_LEVEL_ALL).Infof("found n (%d) filters by [%s]", len(filters), clientId)
    modifier.SetFilters(filters)
    controller.modifiers = append(controller.modifiers, modifier)
    
    controller.accessLock.Unlock()
    
    return modifier
}

func (controller *Controller) DeleteModifier(modifier *Modifier) {
    if (modifier != nil) {
        found := false
        
        controller.accessLock.Lock()
        for i := 0; (i < len(controller.modifiers)) && !found; i++ {
            found = (controller.modifiers[i] == modifier)
            if (found) {
                controller.modifiers = append(controller.modifiers[:i], controller.modifiers[i+1:]...) 
            }
        }
        controller.accessLock.Unlock()
    }
}

func (controller *Controller) SetFilters(clientId string, filters FilterGuid2CompoundFilterMap) error {
    glog.V(GLOG_LEVEL_ALL).Infof("set filters by [%s]", clientId)
    
    controller.accessLock.Lock()
    
    for i := 0; i < len(controller.modifiers); i++ {
        current := controller.modifiers[i]
        if (strings.EqualFold(current.GetClientId(), clientId)) {
            current.SetFilters(filters)
        }
    }
    controller.attachedFilters[clientId] = filters
    
    controller.accessLock.Unlock()
    
    return nil
}

func (controller *Controller) GetFilters(clientId string) FilterGuid2CompoundFilterMap {    
    controller.accessLock.Lock()
    filters := controller.attachedFilters[clientId]
    controller.accessLock.Unlock()
    
    glog.V(GLOG_LEVEL_ALL).Infof("got filters by [%s], total %d", clientId, len(filters))
    
    return filters
}
