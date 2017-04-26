package filter

import (
	"appscope.net/ca"
	"appscope.net/ews"
	"encoding/json"
	"github.com/golang/glog"
	"net/http"
)

type EwsTrigger struct {
	Type     string
	Key      string
	Regexp   string
	Counting bool
}

type EwsAction struct {
	Type    string
	Payload json.RawMessage
}

type EwsFilter struct {
	Triggers []EwsTrigger
	Actions  []EwsAction
	MaxCount int
}

type EwsCompoundFilter struct {
	RequestFilter  EwsFilter
	ResponseFilter EwsFilter
}

type EwsListRequest struct {
	ClientIP  string
	AuthToken []byte
}

type EwsListResponse struct {
	Filters FilterGuid2CompoundFilterMap
}

type EwsConfigureRequest struct {
	ClientIP  string
	Filters   map[string]EwsCompoundFilter
	AuthToken []byte
}

func RegisterEws(controller *Controller) {
	ews.Register("POST", "/filter/v1/list", controller.EwsList)
	ews.Register("POST", "/filter/v1/configure", controller.EwsConfigure)
}

func (controller *Controller) EwsList(responseWriter http.ResponseWriter, request *http.Request) {
	var bytes []byte
	var ewsRequest EwsListRequest

	ews.SetCORSHeader(responseWriter)

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&ewsRequest)

	if err != nil {
		ews.RenderError(responseWriter, http.StatusInternalServerError, "E_WRONG_ARGS", err.Error())
		return
	}

	_, err = ca.Auth(ewsRequest.AuthToken)
	if err != nil {
		ews.RenderError(responseWriter, http.StatusUnauthorized, "E_CMD_FAIL", err.Error())
		return
	}

	glog.V(GLOG_LEVEL_ALL).Infof("list, clientIP: %v", ewsRequest.ClientIP)

	var response EwsListResponse
	response.Filters = controller.GetFilters(ewsRequest.ClientIP)
	bytes, err = json.Marshal(response)

	if err != nil {
		ews.RenderError(responseWriter, http.StatusInternalServerError, "E_CMD_FAIL", err.Error())
	} else {
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusOK)
		responseWriter.Write(bytes)
	}
}

func (controller *Controller) EwsConfigure(responseWriter http.ResponseWriter, request *http.Request) {
	ews.SetCORSHeader(responseWriter)

	decoder := json.NewDecoder(request.Body)

	var ewsRequest EwsConfigureRequest
	err := decoder.Decode(&ewsRequest)

	if err != nil {
		ews.RenderError(responseWriter, http.StatusInternalServerError, "E_WRONG_ARGS", err.Error())
		return
	}

	_, err = ca.Auth(ewsRequest.AuthToken)
	if err != nil {
		ews.RenderError(responseWriter, http.StatusUnauthorized, "E_CMD_FAIL", err.Error())
		return
	}

	glog.V(GLOG_LEVEL_ALL).Infof("configure, clientIP: %v", ewsRequest.ClientIP)

	filters := make(FilterGuid2CompoundFilterMap)

	for key, value := range ewsRequest.Filters {
		var requestActions, responseActions []Action

		glog.Infof("Compound filter[%v]:", key)
		glog.Infof("Request: MaxCount %v", value.RequestFilter.MaxCount)
		glog.Infof("Triggers: %v", value.RequestFilter.Triggers)

		requestActions, err = controller.EwsUnmarshalFilterActions(value.RequestFilter)

		glog.Infof("Response: MaxCount %v", value.ResponseFilter.MaxCount)
		glog.Infof("Triggers: %v", value.ResponseFilter.Triggers)

		if err == nil {
			responseActions, err = controller.EwsUnmarshalFilterActions(value.ResponseFilter)
		} else {
			glog.Warningf("unmarshal response filter[%v] trigger failed: %v", key, err)
		}

		if err == nil {
			requestFilter := NewFilter(controller.EwsTriggersToTriggers(value.RequestFilter.Triggers), requestActions, value.RequestFilter.MaxCount)
			responseFilter := NewFilter(controller.EwsTriggersToTriggers(value.ResponseFilter.Triggers), responseActions, value.ResponseFilter.MaxCount)

			filters[key] = *NewCompoundFilter(requestFilter, responseFilter)

		} else {
			glog.Warningf("unmarshal response filter[%v] action failed: %v", key, err)
		}
	}

	controller.SetFilters(ewsRequest.ClientIP, filters)

	ews.RenderOk(responseWriter, "Filters set")
}

func (controller *Controller) EwsUnmarshalFilterActions(filter EwsFilter) (actions []Action, err error) {
	actions = nil
	err = nil

	//glog.V(GLOG_LEVEL_ALL).Infof("incoming filter actions: %v", filter.Actions)

	for i := 0; i < len(filter.Actions); i++ {
		var action Action = nil

		switch filter.Actions[i].Type {
		case ACTION_DELAY:
			actionImpl := NewActionDelay(0)
			err = json.Unmarshal(filter.Actions[i].Payload, &actionImpl.Payload)
			glog.Infof("Action: %v", *actionImpl)
			action = actionImpl

		case ACTION_PASSTHROUGH:
			actionImpl := NewActionPassthrough()
			err = json.Unmarshal(filter.Actions[i].Payload, &actionImpl.Payload)
			glog.Infof("Action: %v", *actionImpl)
			action = actionImpl

		case ACTION_REPLACE_HEADER:
			actionImpl := NewActionReplaceHeader(nil)
			err = json.Unmarshal(filter.Actions[i].Payload, &actionImpl.Payload)
			glog.Infof("Action: %v", *actionImpl)
			action = actionImpl

		case ACTION_REPLACE_BODY:
			actionImpl := NewActionReplaceBody()
			err = json.Unmarshal(filter.Actions[i].Payload, &actionImpl.Payload)
			glog.Infof("Action: %v", *actionImpl)
			action = actionImpl

		case ACTION_REPLACE_RESPONSE_STATUS:
			actionImpl := NewActionReplaceResponseStatus()
			err = json.Unmarshal(filter.Actions[i].Payload, &actionImpl.Payload)
			glog.Infof("Action: %v", *actionImpl)
			action = actionImpl

		case ACTION_TCP_ERROR:
			actionImpl := NewActionTcpError()
			err = json.Unmarshal(filter.Actions[i].Payload, &actionImpl.Payload)
			glog.Infof("Action: %v", *actionImpl)
			action = actionImpl

		case ACTION_SYNTHETIC_RESPONSE:
			actionImpl := NewActionSyntheticResponse()
			err = json.Unmarshal(filter.Actions[i].Payload, &actionImpl.Payload)
			glog.Infof("Action: %v", *actionImpl)
			action = actionImpl
		}

		if action != nil {
			actions = append(actions, action)
		}
	}

	return actions, err
}

func (controller *Controller) EwsTriggersToTriggers(ewsTriggers []EwsTrigger) (triggers []Trigger) {
	triggers = nil

	for i := 0; i < len(ewsTriggers); i++ {
		triggers = append(triggers, *NewTrigger(ewsTriggers[i].Type, ewsTriggers[i].Key, ewsTriggers[i].Regexp, ewsTriggers[i].Counting))
	}

	return triggers
}
