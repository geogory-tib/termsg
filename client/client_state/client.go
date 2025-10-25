package clientstate

import (
	"client/types"
	"net"
)

type Client struct {
	Inbox       map[string][]types.Message
	server_conn net.Conn
}
