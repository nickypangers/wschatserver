package server

// import (
// 	"log"

// 	"github.com/gorilla/websocket"
// )

// type Message struct {
// 	MessageType int    `json:"messageType"`
// 	Message     string `json:"message"`
// }

// func reader(conn *websocket.Conn) {

// 	log.Println(conn.RemoteAddr())

// 	client := Client{
// 		Address:    conn.RemoteAddr().String(),
// 		Connection: conn,
// 		h:          &hub,
// 	}

// 	hub.Connections[&client] = true
// 	for {
// 		message := &Message{}

// 		if err := conn.ReadJSON(message); err != nil {
// 			log.Println(err)
// 			if hub.Connections[&connection] {
// 				hub.Connections[&connection] = false
// 			}
// 			conn.Close()
// 			return
// 		}

// 		log.Println(hub.Connections[&connection])

// 		var response = Message{MessageType: 1, Message: "Server received"}

// 		if err := conn.WriteJSON(response); err != nil {
// 			log.Println(err)
// 			return
// 		}
// 	}
// }
