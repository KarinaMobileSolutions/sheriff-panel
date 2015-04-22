package main

import (
	auth "github.com/abbot/go-http-auth"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *connection) readPump() {
	for {
		if _, _, err := c.ws.ReadMessage(); err != nil {
			return
		}
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func StartServer() {
	go h.run()

	htpasswd := auth.HtpasswdFileProvider(conf.PasswdFile)
	authenticator := auth.NewBasicAuthenticator("You shall not pass!", htpasswd)

	r := mux.NewRouter()
	r.HandleFunc("/", authenticator.Wrap(serveHome))
	r.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, conf.StaticDir+"/"+r.URL.Path[len("/static/"):])
	})
	r.HandleFunc("/scripts", authenticator.Wrap(serveScripts))
	r.HandleFunc("/scripts/{script}", authenticator.Wrap(getScriptInfo))
	r.HandleFunc("/scripts/chart/{script}", authenticator.Wrap(getScriptChart))
	r.HandleFunc("/ws", serveWebSocket)
	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		errHandler(err, "fatal")
	}
}
