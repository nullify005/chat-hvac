package receiver

import (
	"log"
	"os"
	"regexp"

	"github.com/nullify005/chat-hvac/pkg/adapter"
	"github.com/nullify005/chat-hvac/pkg/hvac"
)

type Receiver struct {
	logger     *log.Logger
	adapter    adapter.Adapter
	shutdown   chan bool
	signatures []ReceiverSignature
	hvac       *hvac.Hvac
}

// definition for what string to match on & then what action to take
type ReceiverSignature struct {
	signature *regexp.Regexp
	handler   ReceiverHandler
}

// generic defintion of a message handler once it's been matched against
type ReceiverHandler func(r *Receiver, s *regexp.Regexp, e *adapter.Event)

type ReceiverOption func(r *Receiver)

func WithLogger(l *log.Logger) ReceiverOption {
	return func(r *Receiver) {
		r.logger = l
	}
}

func WithHvac(h *hvac.Hvac) ReceiverOption {
	return func(r *Receiver) {
		r.hvac = h
	}
}

// Create a new Receiver with default options
// TODO: refactor the hvac requirement & instead register custom handlers which have the logic present within them
func New(a adapter.Adapter, opts ...ReceiverOption) Receiver {
	r := &Receiver{
		adapter:    a,
		logger:     log.New(os.Stdout, "Receiver: ", log.Ldate|log.Ltime|log.Lshortfile),
		shutdown:   make(chan bool, 1),
		signatures: defaultSignatures(),
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.hvac == nil {
		r.hvac = hvac.New()
	}
	return *r
}

// Start up the Listener & Receive events from it
// Receive() is a blocking call which will only exist on a signal or
// shutdown command via the Message
func (r *Receiver) Receive() {
	r.logger.Print("launching event listener")
	recv := make(chan adapter.Event)
	r.adapter.Listen(recv)
	go func() {
		r.logger.Print("starting receiver")
		for {
			evt := <-recv
			r.logger.Printf("received event: %v", evt)
			handled := false
			for _, sig := range r.signatures {
				if sig.signature.Match([]byte(evt.Message)) {
					sig.handler(r, sig.signature, &evt)
					handled = true
					break
				}
			}
			if !handled {
				r.logger.Printf("ignored unhandled event: %v", evt)
			}
		}
	}()
	r.logger.Print("awaiting shutdown signal|command")
	<-r.shutdown
}

// shutdown the receiver & listener loop
func (r *Receiver) Shutdown() {
	r.logger.Print("shutting down")
	r.adapter.Shutdown()
	r.shutdown <- true
}

// registers a new ReceiverSignature with the Reciever to iterate over
// for when the bot is called. The order matters, the 1st registered
// signature will be processed 1st etc.
func (r *Receiver) RegisterSignature(s ReceiverSignature) {
	r.logger.Printf("registering signature: %s", s.signature.String())
	r.signatures = append(r.signatures, s)
}
