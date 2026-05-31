package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/krishnassh/picostatus/internal/storage"
)

type Server struct {
	repo *storage.Repository
	mux  *http.ServeMux
}

func New(repo *storage.Repository) *Server {
	s := &Server{
		repo: repo,
		mux:  http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /api/checks/{id}/history", s.handleHistory)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	checks, err := s.repo.GetChecks()
	if err != nil {
		jsonError(w, "failed to load checks", http.StatusInternalServerError)
		return
	}

	latest, err := s.repo.GetLatestResults()
	if err != nil {
		jsonError(w, "failed to load results", http.StatusInternalServerError)
		return
	}

	uptimes, err := s.repo.GetAllUptimes()
	if err != nil {
		jsonError(w, "failed to load uptime stats", http.StatusInternalServerError)
		return
	}

	resultByCheck := make(map[int64]storage.Result, len(latest))
	for _, res := range latest {
		resultByCheck[res.CheckID] = res
	}

	type checkStatus struct {
		ID        int64          `json:"id"`
		Name      string         `json:"name"`
		Target    string         `json:"target"`
		Status    storage.Status `json:"status"`
		LatencyMS int64          `json:"latency_ms"`
		Error     string         `json:"error,omitempty"`
		Uptime1h  float64        `json:"uptime_1h"`
		Uptime24h float64        `json:"uptime_24h"`
		Uptime7d  float64        `json:"uptime_7d"`
	}

	out := make([]checkStatus, 0, len(checks))
	for _, c := range checks {
		cs := checkStatus{
			ID:     c.ID,
			Name:   c.Name,
			Target: c.Target,
			Status: storage.StatusUnknown,
		}
		if res, ok := resultByCheck[c.ID]; ok {
			cs.Status = res.Status
			cs.LatencyMS = res.LatencyMS
			cs.Error = res.Error
		}
		if u, ok := uptimes[c.ID]; ok {
			cs.Uptime1h = u.Uptime1h
			cs.Uptime24h = u.Uptime24h
			cs.Uptime7d = u.Uptime7d
		}
		out = append(out, cs)
	}

	jsonOK(w, out)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid check id", http.StatusBadRequest)
		return
	}

	results, err := s.repo.GetCheckHistory(id, 30)
	if err != nil {
		jsonError(w, "failed to load history", http.StatusInternalServerError)
		return
	}

	type point struct {
		LatencyMS int64          `json:"latency_ms"`
		Status    storage.Status `json:"status"`
		CheckedAt string         `json:"checked_at"`
	}

	out := make([]point, len(results))
	for i, res := range results {
		out[i] = point{
			LatencyMS: res.LatencyMS,
			Status:    res.Status,
			CheckedAt: res.CreatedAt.Format("15:04:05"),
		}
	}

	jsonOK(w, out)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
