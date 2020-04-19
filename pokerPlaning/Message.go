package pokerplan

// Message struct for sending message to clients
type OutputMessage struct {
	UserName string `json:"userName"`
	Body     string `json:"messageBody"`
}

type InputCommand struct {
	Command string `json:"command"`
}

type StartVoteData struct {
	TopicName string `json:"topic"`
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
