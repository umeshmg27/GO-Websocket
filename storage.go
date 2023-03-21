package chat

import (
	"sync"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type Storage struct {
	clients  []*Client
	mutex    sync.Mutex
	db       driver.Database
	messages driver.Collection
}

func NewStorage() *Storage {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		panic(err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("root", ""),
	})
	if err != nil {
		panic(err)
	}

	db, err := client.Database(nil, "chat")
	if err != nil {
		panic(err)
	}

	messages, err := db.Collection(nil, "messages")
	if err != nil {
		panic(err)
	}

	return &Storage{make([]*Client, 0), sync.Mutex{}, db, messages}
}

func (s *Storage) AddClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients = append(s.clients, client)
}

func (s *Storage) RemoveClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
}

func (s *Storage) Broadcast(message *Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, client := range s.clients {
		if message.Type == MessageTypePrivate {
			if client.GetUsername() == message.RecipientID ||
				client.GetUsername() == message.SenderID {
				client.Send(message)
			}
		} else {
			client.Send(message)
		}
	}
}

func (s *Storage) AddMessage(message *Message) error {
	_, err := s.messages.CreateDocument(nil, message)
	return err
}

func (s *Storage) GetMessages() ([]*Message, error) {
	query := "FOR m IN messages SORT m._key RETURN m"
	cursor, err := s.db.Query(nil, query, nil)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	messages := make([]*Message, 0)
	for {
		var message Message
		_, err := cursor.ReadDocument(nil, &message)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	return messages, nil
}
