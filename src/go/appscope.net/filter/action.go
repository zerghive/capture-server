package filter

import (
    "io"
    "net/http"
)

const ( 
    ACTION_PASSTHROUGH              = "Passthrough"
    ACTION_DELAY                    = "Delay"
    ACTION_REPLACE_HEADER           = "ReplaceHeader"
    ACTION_REPLACE_BODY             = "ReplaceBody"
    ACTION_REPLACE_RESPONSE_STATUS  = "ReplaceStatus"
    ACTION_TCP_ERROR                = "TcpError"
    ACTION_SYNTHETIC_RESPONSE       = "SyntheticResponse"
)

type Action interface {
    GetType() string
    UpdateRequest(request *http.Request) error
    UpdateResponse(response *http.Response) error
}

type KeyValue struct {
    Key   string
    Value string
}

var StatusStrings = map[int]string {
    100: "Continue",
    101: "Switching Protocols",
    200: "OK",
    201: "Created",
    202: "Accepted",
    203: "Non-Authoritative Information",
    204: "No Content",
    205: "Reset Content",
    206: "Partial Content",
    300: "Multiple Choices",
    301: "Moved Permanently",
    302: "Found",
    303: "See Other",
    304: "Not Modified",
    305: "Use Proxy",
    307: "Temporary Redirect",
    400: "Bad Request",
    401: "Unauthorized",
    402: "Payment Required",
    403: "Forbidden",
    404: "Not Found",
    405: "Method Not Allowed",
    406: "Not Acceptable",
    407: "Proxy Authentication Required",
    408: "Request Time-out",
    409: "Conflict",
    410: "Gone",
    411: "Length Required",
    412: "Precondition Failed",
    413: "Request Entity Too Large",
    414: "Request-URI Too Large",
    415: "Unsupported Media Type",
    416: "Requested range not satisfiable",
    417: "Expectation Failed",
    500: "Internal Server Error",
    501: "Not Implemented",
    502: "Bad Gateway",
    503: "Service Unavailable",
    504: "Gateway Time-out",
    505: "HTTP Version not supported",
}

func IsSynthetic(action Action) bool {
    return (action.GetType() == ACTION_SYNTHETIC_RESPONSE)
}

func WriteRequest(request *http.Request, writer io.Writer) error {
    UpdateUserAgent(request)
    
    return request.Write(writer)
}

func WriteResponse(response *http.Response, writer io.Writer) error {
    return response.Write(writer)
}

func UpdateUserAgent(request *http.Request) error {
    if request.Header == nil {
        request.Header = make(http.Header)
    }
    
    if agent := request.Header["User-Agent"]; len(agent) == 0 {
        request.Header.Set("User-Agent", "")
    }

    return nil
}
