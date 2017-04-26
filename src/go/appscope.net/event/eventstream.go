package event

import (
	"github.com/google/gopacket"
)

type Event interface {
	Type() string
}

type EventStream interface {
	PostClientEvent(client gopacket.Endpoint, eventSource func() Event)
}
