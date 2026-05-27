package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

type Server struct {
	store *runtime.Store
}

func NewServer(store *runtime.Store) *Server {
	return &Server{store: store}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/status", s.handleStatus)
	mux.HandleFunc("GET /v1/sessions", s.handleListSessions)
	mux.HandleFunc("POST /v1/sessions", s.handleCreateSession)
	mux.HandleFunc("GET /v1/sessions/{id}/branches", s.handleListBranches)
	mux.HandleFunc("POST /v1/sessions/{id}/branches", s.handleCreateBranch)
	mux.HandleFunc("GET /v1/events", s.handleListEvents)
	mux.HandleFunc("POST /v1/events", s.handleEmitEvent)
	mux.HandleFunc("GET /v1/memory", s.handleListMemory)
	return mux
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	st, err := s.store.Status()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	metrics, _ := s.store.CollectMetrics()
	writeJSON(w, http.StatusOK, map[string]any{"status": st, "metrics": metrics})
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.store.ListSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name      string `json:"name"`
		ProductID string `json:"product_id"`
		FlowID    string `json:"flow_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	sess, err := s.store.CreateSession(body.Name, body.ProductID, body.FlowID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, sess)
}

func (s *Server) handleListBranches(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	branches, err := s.store.ListBranches(sessionID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"branches": branches})
}

func (s *Server) handleCreateBranch(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	var body struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Parent string `json:"parent_branch_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	b, err := s.store.CreateBranch(sessionID, body.Name, runtime.BranchType(body.Type), body.Parent)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	limit := 30
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	events, err := s.store.ListEvents(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

func (s *Server) handleEmitEvent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type      string         `json:"type"`
		SessionID string         `json:"session_id"`
		FlowID    string         `json:"flow_id"`
		Source    string         `json:"source"`
		Payload   map[string]any `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Type == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type required"})
		return
	}
	src := body.Source
	if src == "" {
		src = "api"
	}
	ev, err := s.store.EmitEvent(body.Type, src, body.SessionID, body.FlowID, body.Payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, ev)
}

func (s *Server) handleListMemory(w http.ResponseWriter, r *http.Request) {
	scope := runtime.MemoryScope(r.URL.Query().Get("scope"))
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	eng := memory.NewEngine(s.store)
	if scope != "" {
		entries, err := eng.Retrieve(scope, nil, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"memory": entries})
		return
	}
	entries, err := s.store.ListMemory("", limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"memory": entries})
}

// sessionIDFromPath supports legacy path parsing if needed.
func sessionIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
