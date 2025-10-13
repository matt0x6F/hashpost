package pds

import (
	"encoding/json"
	"net/http"
)

// handleResolveHandle handles DID/handle resolution
func (s *Server) handleResolveHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handle := r.URL.Query().Get("handle")
	if handle == "" {
		http.Error(w, "handle parameter required", http.StatusBadRequest)
		return
	}

	// Validate handle format first
	validator := NewAtprotoValidator()
	if err := validator.ValidateHandle(handle); err != nil {
		s.logger.Error("Invalid handle format", "error", err, "handle", handle)
		http.Error(w, "Invalid handle format", http.StatusBadRequest)
		return
	}

	s.logger.Debug("Resolving handle", "handle", handle)

	// Use DID-based handle resolution
	did, err := s.authService.ResolveHandle(r.Context(), handle)
	if err != nil {
		s.logger.Error("Failed to resolve handle", "error", err, "handle", handle)
		http.Error(w, "Handle not found", http.StatusNotFound)
		return
	}

	// Return the resolved DID
	response := map[string]interface{}{
		"did": did,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
