package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/par1ram/silence/rpc/notifications/internal/domain"
	"github.com/par1ram/silence/rpc/notifications/internal/services"
)

// Server структура HTTP сервера
type Server struct {
	Dispatcher *services.DispatcherService
	Addr       string
}

func NewServer(addr string, dispatcher *services.DispatcherService) *Server {
	return &Server{
		Dispatcher: dispatcher,
		Addr:       addr,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/healthz", s.handleHealthz)
	http.HandleFunc("/notifications", s.handleNotifications)
	log.Printf("[http] server listening on %s", s.Addr)
	return http.ListenAndServe(s.Addr, nil)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req domain.Notification
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
		return
	}
	if req.Type == "" || len(req.Recipients) == 0 || len(req.Channels) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing required fields: type, recipients, channels"))
		return
	}
	if req.ID == "" {
		req.ID = time.Now().Format("20060102T150405.000000000")
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}
	ctx := r.Context()
	if err := s.Dispatcher.Dispatch(ctx, &req); err != nil {
		log.Printf("[http] dispatch error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("dispatch error"))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok"))
}
