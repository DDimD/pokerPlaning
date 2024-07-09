package pokerplan

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	clients           sync.Map //map[string]*Client
	voteList          map[string]*OutputVote
	voteResult        float64
	currentTopicName  string
	voteStoped        bool
}

// NewServer create server object
func NewServer(pattern string) *Server {
	return &Server{
		pattern:           pattern,
		vote:              make(chan *OutputVote, 20),
		voteResultMessage: make(chan bool, 5),
		startVote:         make(chan StartVoteData),
		addClient:         make(chan *Client, 10),
		removeClient:      make(chan *Client, 10),
		clients:           sync.Map{},
		voteList:          make(map[string]*OutputVote),
		voteResult:        0,
		currentTopicName:  "",
		voteStoped:        true,
	}
}

// ConnectHandler handles client connection by websocket
func (srv *Server) connectHandler(w http.ResponseWriter, r *http.Request) {
	//read username from request
	err := r.ParseForm()
	if err != nil {
		log.Printf("connectHandler ParseForm err: %v", err)
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

	ip := strings.Split(r.RemoteAddr, ":")[0]
	clientRaw, ok := srv.clients.Load(ip)
	var client *Client

	var username string
	var role Role
	if !ok {
		username = r.FormValue("username")
		roleInt, _ := strconv.Atoi(r.FormValue("role"))
		role = (Role)(roleInt)

		usernameValidation := regexp.MustCompile(`^[A-zА-я -]*$`)
		usernameValid := usernameValidation.Match([]byte(username))

		//check valid value username parameter
		if len(username) < 1 || !usernameValid {
			log.Println("wrong username param")
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
	} else {
		client = clientRaw.(*Client)

		username = client.Name
		role = client.Role
	}
	client = NewClient(username, role, webSocket, srv, ip)
	client.Listen()
	srv.addClient <- client
	go srv.clients.Range(func(key, value any) bool {
		srvclient := value.(*Client)
		if client.Name == srvclient.Name {
			return true
		}

		clientMessage := &ConnectClientMessage{
			"connect",
			srvclient.Name,
			getRoleDescription(srvclient.Role),
			srvclient.Online,
		}

		client.Send(clientMessage)
		return true
	})

	if !srv.voteStoped {
		client.Send(&voteStartedMessage{"voteStart"})
	}

}

/*
Listen creates handlers to verify that a username is busy,
connect a client by socket, handle adding and deleting clients,

	forwarding an incoming message to other clients
*/
func (srv *Server) Listen() {
	http.HandleFunc("/checkUserName", srv.checkUsernameHandler)
	http.HandleFunc("/whoami", srv.whoamiHandler)
	http.HandleFunc(srv.pattern, srv.connectHandler)

	for {
		select {
		case client := <-srv.addClient:
			srv.clients.Store(client.IP, client)

			log.Printf("user %s connected", client.Name)
			srv.connectClient(client)

		case client := <-srv.removeClient:
			client.Online = false
			srv.clients.Store(client.IP, client)
			srv.disconnectClient(client)
			close(client.done)
			log.Printf("client %s disconnected", client.Name)
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

	srv.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		client.Send(&VotedUser{
			"votedUser",
			vote.UserName,
		})
		return true
	})
}

func (srv *Server) lenDevClients() int {
	cnt := 0
	srv.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		if client.Role != Observer && client.Online == true {
			cnt++
		}
		return true
	})
	return cnt
}

func (srv *Server) connectClient(newClient *Client) {
	srv.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		client.Send(&ConnectClientMessage{
			"connect",
			newClient.Name,
			getRoleDescription(newClient.Role),
			newClient.Online,
		})
		return true
	})
}

func (srv *Server) disconnectClient(rmClient *Client) {
	srv.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		client.Send(&DisconectClientMessage{
			"disconnect",
			rmClient.Name,
		})

		return true
	})

	delete(srv.voteList, rmClient.Name)
	rmClient.webSocket.Close()
}

func (srv *Server) voteStarted() {
	srv.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		client.Send(&voteStartedMessage{
			"voteStart",
		})

		return true
	})
}

func (srv *Server) calculateResult() {
	var sum = 0.
	var cnt = 0.

	for _, vote := range srv.voteList {
		if vote.Vote.IsCoffeeBreak || vote.Vote.IsQuestionMark {
			continue
		}
		sum += vote.Vote.Value
		cnt++
	}

	if cnt == 0. {
		cnt = 1.
	}
	srv.voteResult = sum / cnt
	fmt.Println(cnt)
}

func (srv *Server) sendVoteResultToAll() {
	srv.clients.Range(func(key, value any) bool {
		client := value.(*Client)

		client.Send(&VoteResultMessage{
			srv.voteResult,
			srv.currentTopicName,
			srv.voteList,
		})
		return true
	})

	srv.voteStoped = true
}

// checkUsernameHandler ajax request to verify the use of the given username
func (srv *Server) checkUsernameHandler(rw http.ResponseWriter, req *http.Request) {
	message, err := io.ReadAll(req.Body)
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

type whoamiResponse struct {
	Name string
	Role Role
}

func (srv *Server) whoamiHandler(rw http.ResponseWriter, req *http.Request) {
	ip := strings.Split(req.RemoteAddr, ":")[0]
	clientAny, ok := srv.clients.Load(ip)

	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	client := clientAny.(*Client)

	respRaw := &whoamiResponse{
		Name: client.Name,
		Role: client.Role,
	}
	resp, err := json.Marshal(respRaw)

	if err != nil {
		log.Printf("whoamiHandler err %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusFound)
	rw.Write(resp)
	return

}

// UserExist check user exist in server clients map
func (srv *Server) userExist(username string) bool {
	var exists bool

	srv.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		if client.Name == username {
			exists = true
			return false
		}
		return true
	})

	return exists
}

// GetTopicName returns current topic name
func (srv *Server) GetTopicName() string {
	return srv.currentTopicName
}

func getRoleDescription(role Role) string {
	switch role {
	case 1:
		return "Observer"
	case 2:
		return "Developer"
	case 3:
		return "Maintainer"
	default:
		return "XZ"
	}
}
