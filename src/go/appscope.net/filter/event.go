package filter

import (
)

type TriggerInfo struct {
    Key           string
    OriginalValue string
    RegexpedValue string
    Counting      bool
}

type FilterInfo struct {
    Guid            string
    Count           int
    TriggeredValues []TriggerInfo
}

type HttpRequestEvent struct {
    Info       FilterInfo
    RequestId  int
    HttpConnId uint64
}

type HttpResponseEvent struct {
    Info       FilterInfo
    RequestId  int
    HttpConnId uint64
}

func (t *HttpRequestEvent) Type() string  { return "filter_request" }
func (t *HttpResponseEvent) Type() string { return "filter_response" }
