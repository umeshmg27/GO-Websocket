package chat

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	username string
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn, ""}
}

func (c *Client) Receive() (*Message, error) {
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	var message Message
	err = json.Unmarshal(msg, &message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (c *Client) Send(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) SendErrorMessage(msg string) {
	error := Message{Type: MessageTypeError, Text: msg}
	c.Send(&error)
}

func (c *Client) SetUsername(username string) {
	c.username = username
}

func (c *Client) GetUsername() string {
	return c.username
}
