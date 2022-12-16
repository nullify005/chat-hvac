package adapter

type Adapter interface {
	Listen(chan Event)
	Shutdown()
	Say(Message)
}

type Event struct {
	User      string
	Message   string
	Type      string
	Channel   string
	Timestamp string
}

type Message struct {
	Text      string
	Channel   string
	Timestamp string
	Threaded  bool
}

const (
	AppMentionEvent string = "app_mention"
)
