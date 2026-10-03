package serverx

import (
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Status     string `json:"status"`
	Name       string `json:"name"`
	Production bool   `json:"production"`
	Error      string `json:"error,omitempty"`
}

func (s *Server) statusHandler(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	healthy := s.healthy
	detail := s.detail
	s.mu.Unlock()

	resp := statusResponse{
		Name:       s.cfg.Name,
		Production: s.cfg.Production,
	}

	w.Header().Set("Content-Type", "application/json")

	if !healthy {
		resp.Status = "unavailable"
		resp.Error = detail
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(resp)

		return
	}

	resp.Status = "ok"
	_ = json.NewEncoder(w).Encode(resp)
}
