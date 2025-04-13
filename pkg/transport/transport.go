package transport

import "mcp_server/pkg/messages"

type Transport interface {
	Start() error
	Send(message messages.Request) error
	OnMessage(callback func(message messages.Response))
}
