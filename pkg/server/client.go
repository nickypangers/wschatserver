package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	Address string
}

type ChatMessage struct {
	MessageType int             `json:"messageType"`
	Address     string          `json:"address"`
	Message     string          `json:"message"`
	Data        CommandResponse `json:"data,omitempty"`
	Time        time.Time       `json:"time"`
}

type CommandResponse struct {
	Command string `json:"command"`
	Data    string `json:"data"`
}

const (
	Command      int = 0
	Message      int = 1
	Announcement int = 2
)

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		isCommand := isStringCommand(message)
		if isCommand {
			address := c.Address

			status := c.processCommand(string(message))

			if !status {
				continue
			}

			response := ChatMessage{MessageType: Command, Address: address, Message: "Successfully changed address", Data: CommandResponse{Command: ChangeAddress, Data: c.Address}, Time: time.Now()}

			log.Println(response)

			buf, err := json.Marshal(response)
			if err != nil {
				log.Println(err)
			} else {
				c.hub.broadcast <- []byte(buf)
			}
			continue
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		chatMessage := ChatMessage{MessageType: Message, Address: c.Address, Message: string(message), Time: time.Now()}

		log.Println(chatMessage)

		buf, err := json.Marshal(chatMessage)
		if err != nil {
			log.Println(err)
		} else {
			c.hub.broadcast <- buf
		}

	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	address := params.Get("address")

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), Address: address}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func isStringCommand(message []byte) bool {
	firstChar := string(message[0])
	if firstChar != "/" {
		log.Println("is message")
		return false
	}
	log.Println("is command")
	return true

}
