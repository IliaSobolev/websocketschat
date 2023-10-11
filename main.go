package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clientMutex sync.RWMutex

func main() {
	r := mux.NewRouter()
	conns := make(map[string][]*websocket.Conn, 1)
	r.HandleFunc("/chats/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Debug().Err(err).Str("ip", r.RemoteAddr).Send()
		}
		clientMutex.Lock()
		if conns[key] == nil {
			conns[key] = make([]*websocket.Conn, 0)
		}
		conns[key] = append(conns[key], conn)
		clientMutex.Unlock()

		roomMembers := conns[key]
		defer func() {
			clientMutex.Lock()
			for i, client := range roomMembers {
				if client == conn {
					roomMembers[i] = roomMembers[len(roomMembers)-1]
					roomMembers = roomMembers[:len(conns[key])-1]
				}
			}
			clientMutex.Unlock()
		}()

		log.Info().Str("ip", r.RemoteAddr).Msg("Connected")

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Debug().Err(err).Str("ip", r.RemoteAddr).Send()
				return
			}
			log.Info().Err(err).Str("ip", r.RemoteAddr).Bytes("message", msg).Msg("Message")

			clientMutex.RLock()
			for _, client := range conns[key] {
				err := client.WriteMessage(msgType, msg)
				if err != nil {
					log.Debug().Err(err).Send()
					return
				}
			}
			clientMutex.RUnlock()
		}
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
