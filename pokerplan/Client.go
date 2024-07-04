package pokerplan

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type Role uint8

const (
	Observer   Role = 1
	Developer  Role = 2
	Maintainer Role = 3
)

// Client
type Client struct {
	Name      string
	Role      Role
	IP        string
	Online    bool
	webSocket *websocket.Conn
	server    *Server
	// buffered channel to send messages to the client
	send             chan *VoteResultMessage
	clientConnect    chan *ConnectClientMessage
	clientDisconnect chan *DisconectClientMessage
	voteStarted      chan *voteStartedEvent
}

// NewClient create new client
func NewClient(clientName string, role Role, webSocket *websocket.Conn, server *Server, ip string) *Client {
	if webSocket == nil {
		panic("websocket should be not nil")
	}
	if server == nil {
		panic("server should be not nil")
	}

	return &Client{
		clientName,
		role,
		ip,
		true,
		webSocket,
		server,
		make(chan *VoteResultMessage),
		make(chan *ConnectClientMessage),
		make(chan *DisconectClientMessage),
		make(chan *voteStartedEvent),
	}
}

// Listen function to listen for incoming and outgoing messages
func (cl *Client) Listen() {
	go cl.readCommand()
	cl.sendResults()
}

// readMessage listen, sign and send an incoming message
// to the server channel for broadcast to other clients
func (cl *Client) readCommand() {
	defer func() {
		cl.server.removeClient <- cl
		cl.webSocket.Close()
	}()

	for {
		var cmd InputCommand
		if nil != cl.readJSON(&cmd) {
			break
		}

		switch cmd.Command {
		case "startVote":
			if cl.Role == Maintainer {
				var startData StartVoteData

				if nil != json.Unmarshal(cmd.Message, &startData) {
					break
				}

				cl.server.startVote <- startData
				break
			}
			//TODO: sendError
		case "vote":
			if cl.Role == Observer {
				break
			}

			var vote Vote
			err := json.Unmarshal(cmd.Message, &vote)
			if err != nil {
				fmt.Println(err)
				break
			}

			var outVote OutputVote
			outVote.UserName = cl.Name
			outVote.Vote = vote

			cl.server.vote <- &outVote
		default:
			//TODO: sendError
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
		log.Printf("client %s %v", cl.Name, err)
	}

	return err
}

// sendResults send message to client
func (cl *Client) sendResults() {
	for {
		select {

		case res := <-cl.send:

			err := cl.webSocket.WriteJSON(res)
			if err != nil {
				log.Printf("client %s %v", cl.Name, err)
				break
			}

		case newClient := <-cl.clientConnect:
			cl.webSocket.WriteJSON(newClient)
		case disconnectClient := <-cl.clientDisconnect:
			cl.webSocket.WriteJSON(disconnectClient)
		case voteStarted := <-cl.voteStarted:
			cl.webSocket.WriteJSON(voteStarted)
		}
	}
}
