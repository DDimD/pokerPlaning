package chat

import (
	"log"

	"github.com/gorilla/websocket"
)

type Role uint8

const (
	ProjectManager Role = 1
	Developer
	QA
)

// Client
type Client struct {
	name      string
	role      Role
	webSocket *websocket.Conn
	server    *Server
	// buffered channel to send messages to the client
	send chan *OutputMessage
}

//NewClient create new client
func NewClient(clientName string, role Role, webSocket *websocket.Conn, server *Server) *Client {
	if webSocket == nil {
		panic("websocket should be not nil")
	}
	if server == nil {
		panic("server should be not nil")
	}

	return &Client{
		clientName,
		role,
		webSocket,
		server,
		make(chan *OutputMessage, 25)}
}

//Listen function to listen for incoming and outgoing messages
func (cl *Client) Listen() {
	go cl.readMessage()
	cl.writeMessage()
}

//readMessage listen, sign and send an incoming message
// to the server channel for broadcast to other clients
func (cl *Client) readCommand() {
	defer func() {
		cl.server.removeClient <- cl
		cl.webSocket.Close()
	}()

	for {
		var cmd InputCommand
		if nil == cl.readJSON(&cmd) {
			break
		}

		switch cmd.Command {
		case "startVote":
			if cl.role == ProjectManager {
				var startData StartVoteData

				if nil == cl.readJSON(&startData) {
					break
				}

				cl.server.startVote <- startData
				break
			}
			//TODO: sendError
		case "vote":
			if cl.role == ProjectManager {
				break
			}

			var vote Vote
			if nil == cl.readJSON(&vote) {
				break
			}

			var outVote OutputVote
			outVote.UserName = cl.name
			outVote.Vote = vote

			cl.server.vote <- &outVote
		default:
		}

	}
}

// func (cl *Client) sendError() {
// 	for {
// 		msg := <-cl.send

// 		err := cl.webSocket.WriteJSON(msg)
// 		if err != nil {
// 			log.Printf("client %s %v", cl.name, err)
// 			break
// 		}
// 	}
// }

func (cl *Client) readJSON(val interface{}) error {
	err := cl.webSocket.ReadJSON(&val)

	if err != nil {
		log.Printf("client %s %v", cl.name, err)
	}

	return err
}

// writeMessage send message to client
func (cl *Client) writeMessage() {
	for {
		msg := <-cl.send

		err := cl.webSocket.WriteJSON(msg)
		if err != nil {
			log.Printf("client %s %v", cl.name, err)
			break
		}
	}
}
