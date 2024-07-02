package pokerplan

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// Server manages connected clients and message forwarding
type Server struct {
	pattern           string
	vote              chan *OutputVote
	voteResultMessage chan bool
	startVote         chan StartVoteData
	addClient         chan *Client
	removeClient      chan *Client
	clients           map[string]*Client
	rwMutex           sync.RWMutex
	voteList          map[string]*OutputVote
	voteResult        float64
	currentTopicName  string
	voteStoped        bool
}

// NewServer create server object
func NewServer(pattern string) *Server {
	return &Server{
		pattern,
		make(chan *OutputVote, 20),
		make(chan bool, 5),
		make(chan StartVoteData),
		make(chan *Client, 10),
		make(chan *Client, 10),
		make(map[string]*Client),
		sync.RWMutex{},
		make(map[string]*OutputVote),
		0,
		"",
		true,
	}
}

// ConnectHandler handles client connection by websocket
func (srv *Server) connectHandler(w http.ResponseWriter, r *http.Request) {
	//read username from request
	err := r.ParseForm()
	if err != nil {
		log.Printf("connectHandler ParseForm err: %v", err)
	}

	username := r.FormValue("username")
	roleInt, _ := strconv.Atoi(r.FormValue("role"))
	role := (Role)(roleInt)

	//check valid value username parameter
	if len(username) < 1 {
		log.Println("wrong username param")
		return
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	webSocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket connection upgrade error: %v", err)
		return
	}

	//check username existence
	if srv.userExist(username) {
		webSocket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4018, "username already used"))
		webSocket.Close()
		log.Printf("Username %s already used", username)
		return
	}
	//TODO: need check valid role

	client := NewClient(username, role, webSocket, srv)
	srv.addClient <- client
	for _, srvclient := range srv.clients {
		if client.Name == srvclient.Name {
			continue
		}

		clientMessage := &ConnectClientMessage{
			"connect",
			srvclient.Name,
			getRoleDescription(srvclient.Role),
		}

		webSocket.WriteJSON(clientMessage)
	}

	if !srv.voteStoped {
		webSocket.WriteJSON(&voteStartedEvent{
			"voteStart",
		})
	}

	client.Listen()

}

/*
Listen creates handlers to verify that a username is busy,
connect a client by socket, handle adding and deleting clients,

	forwarding an incoming message to other clients
*/
func (srv *Server) Listen() {
	http.HandleFunc("/checkUserName", srv.checkUsernameHandler)
	http.HandleFunc(srv.pattern, srv.connectHandler)

	for {
		select {
		case client := <-srv.addClient:
			srv.clients[client.Name] = client
			log.Printf("user %s connected", client.Name)
			srv.connectClient(client)

		case client := <-srv.removeClient:
			delete(srv.clients, client.Name)
			log.Printf("client %s removed", client.Name)
			srv.disconnectClient(client)

		case data := <-srv.startVote:
			if srv.voteStoped {
				srv.start(data)
			}
			//TODO: Добавить сообщение, что голосовалка уже началась
		case vote := <-srv.vote:
			if !srv.voteStoped {
				srv.addNewVote(vote)
			}
			//TODO: Добавить сообщение, что голосование закончено
		case tr := <-srv.voteResultMessage:
			fmt.Println(tr)
			srv.calculateResult()
			srv.sendVoteResultToAll()
		}
	}
}

func (srv *Server) start(data StartVoteData) {
	srv.currentTopicName = data.TopicName
	srv.voteList = make(map[string]*OutputVote)
	srv.voteStoped = false
	srv.voteStarted()
}

func (srv *Server) addNewVote(vote *OutputVote) {
	srv.voteList[vote.UserName] = vote
	if len(srv.voteList) >= srv.lenDevClients() {
		srv.voteResultMessage <- true
	}
}

func (srv *Server) lenDevClients() int {
	cnt := 0
	for _, cl := range srv.clients {
		if cl.Role != ProjectManager {
			cnt++
		}
	}
	return cnt
}

func (srv *Server) connectClient(newClient *Client) {
	for _, client := range srv.clients {
		client.clientConnect <- &ConnectClientMessage{
			"connect",
			newClient.Name,
			getRoleDescription(newClient.Role),
		}
	}
}

func (srv *Server) disconnectClient(rmClient *Client) {
	for _, client := range srv.clients {
		client.clientDisconnect <- &DisconectClientMessage{
			"disconnect",
			rmClient.Name,
		}
	}
}

func (srv *Server) voteStarted() {
	for _, client := range srv.clients {
		client.voteStarted <- &voteStartedEvent{
			"voteStart",
		}
	}
}

func (srv *Server) calculateResult() {
	var sum float64 = 0.
	var cnt float64 = 0.

	for _, vote := range srv.voteList {
		if vote.Vote.IsCoffeeBreak || vote.Vote.IsQuestionMark {
			continue
		}
		sum += vote.Vote.Value
		cnt++
	}
	srv.voteResult = sum / cnt
	fmt.Println(cnt)
}

func (srv *Server) sendVoteResultToAll() {
	for _, client := range srv.clients {
		client.send <- &VoteResultMessage{
			srv.voteResult,
			srv.currentTopicName,
			srv.voteList,
		}
	}
	srv.voteStoped = true
}

// func (srv *Server) sendAll(res float64) {
// 	for _, client := range srv.clients {
// 		if client.name != msg.UserName {
// 			client.send <- msg
// 		}
// 	}
// }

// checkUsernameHandler ajax request to verify the use of the given username
func (srv *Server) checkUsernameHandler(rw http.ResponseWriter, req *http.Request) {
	message, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(fmt.Sprintf("username check request read error"))
	}
	username := string(message)

	if srv.userExist(username) {
		rw.WriteHeader(http.StatusTeapot)
		rw.Write([]byte("username already used"))
		return
	}

	rw.Write([]byte("username not used"))
}

// UserExist check user exist in server clients map
func (srv *Server) userExist(username string) bool {
	srv.rwMutex.RLock()
	defer srv.rwMutex.RUnlock()

	_, ok := srv.clients[username]
	return ok
}

// GetClients return all clients in array
func (srv *Server) GetClients() []User {
	users := make([]User, 0, len(srv.clients))
	for _, val := range srv.clients {
		users = append(users, User{
			val.Name,
			getRoleDescription(val.Role),
		})
	}
	return users
}

// GetTopicName returns current topic name
func (srv *Server) GetTopicName() string {
	return srv.currentTopicName
}

func getRoleDescription(role Role) string {
	switch role {
	case 1:
		return "PM"
	case 2:
		return "Developer"
	case 3:
		return "QA"
	default:
		return "XZ"
	}
}
