package receiver

import (
	"regexp"

	"github.com/nullify005/chat-hvac/pkg/adapter"
)

const (
	setSignature      string = "(.+) set (.+) (.+)"
	statusSignature   string = "(.+) (status|state)"
	pingSignature     string = "(.+) (ping|hi|hello)"
	helpSignature     string = "(.+) help"
	shutdownSignature string = "(.+) shutdown"
	defaultSignature  string = "(.+) .*"

	helpReply     string = "I'm expecting something like\n`@hvac (help|status|set|get|shutdown) key [value]`"
	shutdownReply string = "Shutdown command received. Going to sleep now, bye ..."
	defaultReply  string = "I'm not sure what you are after. :shrug:"
)

// returns the default list of signatures which are supported and their associated handlers
func defaultSignatures() []ReceiverSignature {
	return []ReceiverSignature{
		{
			signature: regexp.MustCompile(setSignature),
			handler:   setHandler,
		},
		{
			signature: regexp.MustCompile(statusSignature),
			handler:   statusHandler,
		},
		{
			signature: regexp.MustCompile(helpSignature),
			handler:   helpHandler,
		},
		{
			signature: regexp.MustCompile(pingSignature),
			handler:   pingHandler,
		},
		{
			signature: regexp.MustCompile(shutdownSignature),
			handler:   shutdownHandler,
		},
		{
			signature: regexp.MustCompile(defaultSignature),
			handler:   defaultHandler,
		},
	}
}

// the default handler which catches any mention which didn't get processed by
// another handler
func defaultHandler(r *Receiver, s *regexp.Regexp, e *adapter.Event) {
	m := adapter.Message{
		Text:      defaultReply + "\n" + helpReply,
		Channel:   e.Channel,
		Threaded:  false,
		Timestamp: e.Timestamp,
	}
	r.adapter.Say(m)
}

// sends the help message
func helpHandler(r *Receiver, s *regexp.Regexp, e *adapter.Event) {
	m := adapter.Message{
		Text:      helpReply,
		Channel:   e.Channel,
		Threaded:  false,
		Timestamp: e.Timestamp,
	}
	r.adapter.Say(m)
}

// respond to hello are you there requests
func pingHandler(r *Receiver, s *regexp.Regexp, e *adapter.Event) {
	reply := "Err, not sure how I ended up here in the ping handler to be honest ... :confused:"
	match := s.FindSubmatch([]byte(e.Message))
	switch {
	case string(match[2]) == "ping":
		reply = "pong"
	case string(match[2]) == "hi":
		reply = ":wave:"
	case string(match[2]) == "hello":
		reply = "Yes, I'm listening ..."
	case string(match[2]) == "wave":
		reply = ":wave:"
	}
	m := adapter.Message{
		Text:      reply,
		Channel:   e.Channel,
		Threaded:  false,
		Timestamp: e.Timestamp,
	}
	r.adapter.Say(m)
}

// get the hvac status
func setHandler(r *Receiver, s *regexp.Regexp, e *adapter.Event) {
	match := s.FindSubmatch([]byte(e.Message))
	m := adapter.Message{
		Text:      r.hvac.Set(string(match[2]), string(match[3])),
		Channel:   e.Channel,
		Threaded:  false,
		Timestamp: e.Timestamp,
	}
	r.adapter.Say(m)
}

// receive and process the shutdown command
func shutdownHandler(r *Receiver, s *regexp.Regexp, e *adapter.Event) {
	m := adapter.Message{
		Text:      shutdownReply,
		Channel:   e.Channel,
		Threaded:  false,
		Timestamp: e.Timestamp,
	}
	r.adapter.Say(m)
	r.Shutdown()
}

// get the hvac status
func statusHandler(r *Receiver, s *regexp.Regexp, e *adapter.Event) {
	m := adapter.Message{
		Text:      r.hvac.Status(),
		Channel:   e.Channel,
		Threaded:  false,
		Timestamp: e.Timestamp,
	}
	r.adapter.Say(m)
}
