package console

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nullify005/chat-hvac/pkg/adapter"
)

type Listener struct {
	shutdown chan bool
	logger   *log.Logger
}

type ListenerOption func(l *Listener)

func WithLogger(l *log.Logger) ListenerOption {
	return func(s *Listener) {
		s.logger = l
	}
}

func New(opts ...ListenerOption) *Listener {
	l := &Listener{
		logger:   log.New(os.Stdout, "SlackListener: ", log.Ldate|log.Ltime|log.Lshortfile),
		shutdown: make(chan bool, 1),
	}
	for _, opt := range opts {
		opt(l)
	}
	l.logger.Print("using console adapter")
	return l
}

func (l *Listener) Listen(output chan adapter.Event) {
	go func() {
		l.logger.Print("setting up listener loop")
		scanner := bufio.NewScanner(os.Stdin)
	listener:
		for {
			select {
			case <-l.shutdown:
				log.Print("received close, shutting down")
				close(output)
				break listener
			default:
				scanner.Scan()
				if err := scanner.Err(); err != nil {
					l.logger.Printf("error reading from stdin. cause: %v", err)
					l.Shutdown()
					continue
				}
				evt := &adapter.Event{
					User:      os.Getenv("USER"),
					Type:      adapter.AppMentionEvent,
					Channel:   "os.Stdin",
					Timestamp: fmt.Sprint(time.Now().Unix()),
					Message:   scanner.Text(),
				}
				output <- *evt
			}
		}
		l.logger.Print("listener ending")
	}()

}

func (l *Listener) Say(m adapter.Message) {
	fmt.Printf(">> (%s) %s\n", m.Channel, m.Text)
}

func (l *Listener) Shutdown() {
	l.logger.Print("shutting down")
	l.shutdown <- true
}
