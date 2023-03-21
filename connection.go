package chat

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Connection struct {
	url string
}

func NewConnection(host string, port int) *Connection {
	return &Connection{fmt.Sprintf("ws://%s:%d/ws", host, port)}
}

func (c *Connection) Connect(username string) (*websocket.Conn, error) {
	header := http.Header{}
	header.Set("username", username)
	conn, _, err := websocket.DefaultDialer.Dial(c.url, header)
	return conn, err
}
