package client

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

func NewWsClient(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	upgrader := websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 2 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	return wsConn
}
