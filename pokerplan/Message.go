package pokerplan

import "encoding/json"

// Message struct for sending message to clients
type OutputMessage struct {
	UserName string `json:"userName"`
	Body     string `json:"messageBody"`
}

type InputCommand struct {
	Command string          `json:"command"`
	Message json.RawMessage `json:"body"`
}

type StartVoteData struct {
	TopicName string `json:"topic"`
}
type User struct {
	Name string
	Role string
}

type voteStartedEvent struct {
	Command string
}

type ConnectClientMessage struct {
	Command  string
	UserName string
	Role     string
	Online   bool
}

type DisconectClientMessage struct {
	Command  string
	UserName string
	Ip       string
}

type VoteResultMessage struct {
	Result    float64                `json:"voteResult"`
	TopicName string                 `json:"topic"`
	Votes     map[string]*OutputVote `json:"votes"`
}

type Vote struct {
	Value          float64 `json:"value"`
	IsCoffeeBreak  bool    `json:"isCoffeeBreak"`
	IsQuestionMark bool    `json:"isQuestionMark"`
}

type OutputVote struct {
	UserName string `json:"userName"`
	Vote     Vote   `json:"vote"`
}
