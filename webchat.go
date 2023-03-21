// package webChat

// import (
// 	"log"
// 	"net/http"

// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }

// type hub struct {
// 	rooms      map[string]*Room
// 	register   chan *Client
// 	unregister chan *Client
// 	broadcast  chan *Message
// }

// func newHub() *hub {
// 	return &hub{
// 		rooms:      make(map[string]*Room),
// 		register:   make(chan *Client),
// 		unregister: make(chan *Client),
// 		broadcast:  make(chan *Message),
// 	}
// }

// func (h *hub) run() {
// 	for {
// 		select {
// 		case client := <-h.register:
// 			room, ok := h.rooms[client.roomID]
// 			if !ok {

// 				room = NewRoom(client.roomID, client.room.Name, newArangoDBStore("webchat"))
// 				h.rooms[client.roomID] = room
// 				go room.Run()
// 			}

// 			room.Members[client] = true
// 			log.Printf("Client %s joined room %s\n", client.id, client.roomID)

// 			room.broadcastJoinEvent(client)

// 		case client := <-h.unregister:
// 			if room, ok := h.rooms[client.roomID]; ok {
// 				if _, ok := room.Members[client]; ok {
// 					delete(room.Members, client)
// 					close(client.send)
// 					log.Printf("Client %s left room %s\n", client.id, client.roomID)

// 					room.broadcastLeaveEvent(client)

// 					if membership, err := room.store.GetMembership(client.userID, client.roomID); err == nil && membership != nil {
// 						err := room.store.DeleteMembership(*membership)
// 						if err != nil {
// 							log.Printf("Failed to delete membership: %s\n", err)
// 						}
// 					}

// 					if len(room.Members) == 0 {
// 						delete(h.rooms, client.roomID)
// 						log.Printf("Room %s has been deleted due to no members\n", client.roomID)
// 					}
// 				}
// 			}

// 		case message := <-h.broadcast:
// 			room, ok := h.rooms[message.RoomID]
// 			if ok {
// 				room.broadcastMessage(message)
// 			}
// 		}
// 	}
// }

// type Room struct {
// 	name    string
// 	clients map[*Client]bool
// 	join    chan *Client
// 	leave   chan *Client
// 	msgs    chan ChatMessage
// }

// func NewRoom(name string) *Room {
// 	return &Room{
// 		name:    name,
// 		clients: make(map[*Client]bool),
// 		join:    make(chan *Client),
// 		leave:   make(chan *Client),
// 		msgs:    make(chan ChatMessage),
// 	}
// }

// func (r *Room) Run() {
// 	for {
// 		select {
// 		case client := <-r.join:
// 			r.clients[client] = true
// 			log.Printf("Client %s joined room %s", client.username, r.name)
// 		case client := <-r.leave:
// 			r.unregister(client)
// 		case msg := <-r.msgs:
// 			for client := range r.clients {
// 				select {
// 				case client.send <- msg:
// 				default:
// 					close(client.send)
// 					delete(r.clients, client)
// 				}
// 			}
// 		}
// 	}
// }

// func (r *Room) register(client *Client) {
// 	r.join <- client
// }

// func (r *Room) unregister(client *Client) {
// 	r.leave <- client
// }

// func (r *Room) broadcast(msg ChatMessage, sender *Client) {
// 	r.Lock()
// 	defer r.Unlock()
// 	for client := range r.clients {
// 		if client != sender {
// 			select {
// 			case client.send <- msg:
// 			default:
// 				close(client.send)
// 				delete(r.clients, client)
// 			}
// 		}
// 	}
// }

// type Client struct {
// 	id       string
// 	userID   string
// 	userName string
// 	roomID   string
// 	conn     *websocket.Conn
// 	send     chan *Event
// }

// type Message struct {
// 	UserID  string `json:"userId"`
// 	RoomID  string `json:"roomId"`
// 	Message string `json:"message"`
// 	Time    int64  `json:"time"`
// }

// type User struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// }

// type Event struct {
// 	Type    string   `json:"type"`
// 	User    *User    `json:"user,omitempty"`
// 	Message *Message `json:"message,omitempty"`
// }

// type Store interface {
// 	CreateMembership(membership *Membership) error
// 	GetMembership(userID, roomID string) (*Membership, error)
// 	DeleteMembership(membership Membership) error
// 	CreateMessage(message *Message) error
// 	GetMessages(roomID string) ([]*Message, error)
// }

// type Membership struct {
// 	UserID string `json:"userId"`
// 	RoomID string `json:"roomId"`
// }

// type ArangoDBStore struct {
// 	dbName string
// 	client *driver.Client
// }

// func newArangoDBStore(dbName string) *ArangoDBStore {
// 	conn, err := driver.NewClient(driver.ClientConfig{
// 		Connection: driver.ConnectionConfig{
// 			Endpoint: "http://localhost:8529",
// 		},
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return &ArangoDBStore{
// 		dbName: dbName,
// 		client: conn,
// 	}
// }
// func (s *ArangoDBStore) CreateMembership(membership *Membership) error {
// 	col, err := s.getMembershipCollection()
// 	if err != nil {
// 		return err
// 	}
// 	_, err = col.CreateDocument(nil, membership)
// 	return err

// }

// func (s *ArangoDBStore) GetMembership(userID, roomID string) (*Membership, error) {
// 	col, err := s.getMembershipCollection()
// 	if err != nil {
// 		return nil, err
// 	}
// 	query := "FOR m IN memberships FILTER m.userId == @userId && m.roomId ==@roomId RETURN m"

// 	bindVars := map[string]interface{}{
// 		"userId": userID,
// 		"roomId": roomID,
// 	}

// 	cursor, err := s.client.Query(nil, query, bindVars)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close()

// 	if !cursor.HasMore() {
// 		return nil, nil
// 	}

// 	var membership Membership
// 	_, err = cursor.ReadDocument(nil, &membership)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &membership, nil

// }

// func (s *ArangoDBStore) DeleteMembership(membership Membership) error {
// 	col, err := s.getMembershipCollection()
// 	if err != nil {
// 		return err
// 	}
// 	_, err = col.RemoveDocument(nil, membership.ID)
// 	return err
// }

// func (s *ArangoDBStore) CreateMessage(message *Message) error {
// 	col, err := s.getMessageCollection()
// 	if err != nil {
// 		return err
// 	}
// 	_, err = col.CreateDocument(nil, message)
// 	return err
// }

// func (s *ArangoDBStore) GetMessages(roomID string) ([]*Message, error) {
// 	col, err := s.getMessageCollection()
// 	if err != nil {
// 		return nil, err
// 	}
// 	query := "FOR m IN messages FILTER m.roomId == @roomId SORT m.time RETURN m"

// 	bindVars := map[string]interface{}{
// 		"roomId": roomID,
// 	}

// 	cursor, err := s.client.Query(nil, query, bindVars)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close()

// 	var messages []*Message
// 	for cursor.HasMore() {
// 		var message Message
// 		_, err = cursor.ReadDocument(nil, &message)
// 		if err != nil {
// 			return nil, err
// 		}

// 		messages = append(messages, &message)
// 	}

// 	return messages, nil
// }

// func (s *ArangoDBStore) getMembershipCollection() (driver.Collection, error) {
// 	db, err := s.client.Database(nil, s.dbName)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return db.Collection(nil, "memberships")
// }

// func (s *ArangoDBStore) getMessageCollection() (driver.Collection, error) {
// 	db, err := s.client.Database(nil, s.dbName)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return db.Collection(nil, "messages")
// }

// func init() {
// 	http.HandleFunc("/ws", handleWebSocket)
// }

// func handleWebSocket(w http.ResponseWriter, r *http.Request) {
// 	roomID := r.URL.Query().Get("roomId")
// 	userID := r.URL.Query().Get("userId")
// 	userName := r.URL.Query().Get("userName")
// 	if roomID == "" || userID == "" || userName == "" {
// 		http.Error(w, "Missing required parameters", http.StatusBadRequest)
// 		return
// 	}

// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("Failed to upgrade connection: %s\n", err)
// 		return
// 	}

// 	client := &Client{
// 		id:       uuid.New().String(),
// 		userID:   userID,
// 		userName: userName,
// 		roomID:   roomID,
// 		conn:     conn,
// 		send:     make(chan *Event),
// 	}

// 	hub := getHub(roomID)
// 	hub.register <- client

// 	go client.writePump()
// 	go client.readPump()
// }

// func getHub(roomID string) *Hub {
// 	hubsLock.Lock()
// 	defer hubsLock.Unlock()
// 	hub, ok := hubs[roomID]
// 	if !ok {
// 		hub = newHub()
// 		hubs[roomID] = hub
// 		go hub.run()
// 	}

// 	return hub

// }

// func main() {
// 	// Initialize ArangoDB connection and store
// 	store := NewArangoDBStore("http://localhost:8529", "myDatabase")
// 	// Create a new chat room
// 	room := &Room{
// 		ID:      uuid.New().String(),
// 		Name:    "General",
// 		Members: []string{},
// 	}
// 	err := store.CreateRoom(room)
// 	if err != nil {
// 		log.Fatalf("Failed to create room: %s", err)
// 	}

// 	// Start HTTP server
// 	log.Println("Starting server...")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
