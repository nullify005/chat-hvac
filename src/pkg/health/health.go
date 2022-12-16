package health

import (
	"io"
	"log"
	"net/http"
	"os"
)

type Health struct {
	logger *log.Logger
	listen string
}

type HealthOption func(h *Health)

func WithLogger(l *log.Logger) HealthOption {
	return func(h *Health) {
		h.logger = l
	}
}

func WithListen(l string) HealthOption {
	return func(h *Health) {
		h.listen = l
	}
}

func New(opts ...HealthOption) *Health {
	h := &Health{
		logger: log.New(os.Stdout, "Health: ", log.Ldate|log.Ltime|log.Lshortfile),
		listen: ":8080",
	}
	for _, opt := range opts {
		opt(h)
	}
	h.logger.Print("setting up health handlers")
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/health", healthHandler)
	return h
}

func (h *Health) Run() {
	go func() {
		h.logger.Print("starting health listen and serve")
		h.logger.Fatal(http.ListenAndServe(h.listen, nil))
	}()
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	io.WriteString(w, "not implemented")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "ok")
}
