package chat

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	storage  *Storage
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	storage := NewStorage()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &Server{storage, upgrader}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn)
	s.storage.AddClient(client)

	go s.handleClient(client)
}

func (s *Server) handleClient(client *Client) {
	defer s.storage.RemoveClient(client)

	for {
		msg, err := client.Receive()
		if err != nil {
			log.Println(err)
			break
		}

		switch msg.Type {
		case MessageTypeText:
			s.handleTextMessage(client, msg)
		case MessageTypePrivate:
			s.handlePrivateMessage(client, msg)
		case MessageTypeJoin:
			s.handleJoinMessage(client, msg)
		case MessageTypeLeave:
			s.handleLeaveMessage(client, msg)
		}
	}
}

func (s *Server) handleTextMessage(client *Client, msg *Message) {
	s.storage.BroadcastMessage(client, msg)
}

func (s *Server) handlePrivateMessage(client *Client, msg *Message) {
	recipient := s.storage.GetClientByID(msg.RecipientID)
	if recipient == nil {
		client.SendErrorMessage(fmt.Sprintf("User %s is not online", msg.RecipientID))
		return
	}
	recipient.Send(msg)
}

func (s *Server) handleJoinMessage(client *Client, msg *Message) {
	client.SetUsername(msg.Username)
	s.storage.BroadcastJoinMessage(client)
}

func (s *Server) handleLeaveMessage(client *Client, msg *Message) {
	s.storage.RemoveClient(client)
	s.storage.BroadcastLeaveMessage(client)
}
