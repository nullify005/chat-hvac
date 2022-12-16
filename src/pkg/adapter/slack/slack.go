package slack

import (
	"fmt"
	"log"
	"os"

	"github.com/nullify005/chat-hvac/pkg/adapter"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Listener struct {
	shutdown chan bool
	logger   *log.Logger
	client   *slack.Client
	socket   *socketmode.Client
	output   chan adapter.Event
}

var listener Listener // a mirror of the Listener struct so that we can use the vars within

type ListenerOption func(l *Listener)

func WithLogger(l *log.Logger) ListenerOption {
	return func(s *Listener) {
		s.logger = l
	}
}

func New(botToken, appToken string, opts ...ListenerOption) *Listener {
	l := &Listener{
		logger:   log.New(os.Stdout, "SlackListener: ", log.Ldate|log.Ltime|log.Lshortfile),
		shutdown: make(chan bool, 1),
	}
	for _, opt := range opts {
		opt(l)
	}
	l.logger.Print("using slack adapter")
	l.client = slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
		slack.OptionLog(l.logger),
	)
	l.socket = socketmode.New(
		l.client,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(l.logger),
	)
	listener = *l
	return l
}

func (l *Listener) Listen(output chan adapter.Event) {
	listener.output = output
	handler := socketmode.NewSocketmodeHandler(l.socket)
	handler.Handle(socketmode.EventTypeConnecting, middlewareConnecting)
	handler.Handle(socketmode.EventTypeConnectionError, middlewareConnectionError)
	handler.Handle(socketmode.EventTypeConnected, middlewareConnected)
	handler.HandleEvents(slackevents.AppMention, middlewareAppMentionEvent)
	go func() {
		l.logger.Fatal(handler.RunEventLoop())
	}()
}

func (l *Listener) Say(m adapter.Message) {
	var err error
	if m.Threaded {
		_, _, err = l.client.PostMessage(m.Channel,
			slack.MsgOptionText(m.Text, false),
			slack.MsgOptionTS(fmt.Sprint(m.Timestamp)),
		)
	} else {
		_, _, err = l.client.PostMessage(m.Channel, slack.MsgOptionText(m.Text, false))
	}
	if err != nil {
		l.logger.Printf("unable to post message. cause: %v", err)
	}
}

func (l *Listener) Shutdown() {
	l.logger.Print("shutting down")
	l.shutdown <- true
}

func middlewareConnecting(evt *socketmode.Event, client *socketmode.Client) {
	listener.logger.Print("socketmode connectng")
}

func middlewareConnectionError(evt *socketmode.Event, client *socketmode.Client) {
	listener.logger.Print("socketmode connection error")
}

func middlewareConnected(evt *socketmode.Event, client *socketmode.Client) {
	listener.logger.Print("socketmode connected")
}

func middlewareAppMentionEvent(evt *socketmode.Event, client *socketmode.Client) {
	listener.logger.Print("socketmode AppMentionEvent")
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		listener.logger.Printf("ignored event: %+v", evt)
		return
	}

	client.Ack(*evt.Request)

	ev, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		listener.logger.Printf("ignored event: %+v", ev)
		return
	}
	event := &adapter.Event{
		User:      ev.User,
		Message:   ev.Text,
		Type:      adapter.AppMentionEvent,
		Channel:   ev.Channel,
		Timestamp: ev.EventTimeStamp,
	}
	listener.output <- *event
}
