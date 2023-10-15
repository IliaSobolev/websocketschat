package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type Client struct {
	Connection *websocket.Conn
}

type Message struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

func (cl *Client) Send(message *Message) {
	jsondata, err := json.Marshal(message)
	if err != nil {
		log.Println("marhal", err)
		return
	}
	err = cl.Connection.WriteMessage(websocket.TextMessage, jsondata)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (cl *Client) Handle(handler func(m *Message)) {
	for {
		var message *Message
		messageType, msg, err := cl.Connection.ReadMessage()
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Println("unmarshal", err)
			return
		}
		if messageType == websocket.TextMessage {
			handler(message)
		}
	}
}
func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/chats/room1"}

	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Printf("connected to %s", u.String())
	client := &Client{
		Connection: c,
	}
	go client.Handle(func(m *Message) { fmt.Println(m.Type, m.Content) })
	client.Send(&Message{Type: 12, Content: "aboba"})
	defer c.Close()

	select {}
}
