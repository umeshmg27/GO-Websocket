package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {

	flag.Parse()

	storage := chat.NewStorage()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if err != nil {
			log.Println(err)
			return
		}

		username := r.Header.Get("username")
		if username == "" {
			conn.Close()
			return
		}

		client := chat.NewClient(conn)
		client.SetUsername(username)

		storage.AddClient(client)
		defer storage.RemoveClient(client)

		messages, err := storage.GetMessages()
		if err != nil {
			log.Println(err)
		}

		for _, message := range messages {
			client.Send(message)
		}

		for {
			msg, err := client.Receive()
			if err != nil {
				log.Println(err)
				break
			}

			message := chat.NewMessage(msg, username)

			if err := storage.AddMessage(message); err != nil {
				log.Println(err)
			}

			storage.Broadcast(message)
		}
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}
