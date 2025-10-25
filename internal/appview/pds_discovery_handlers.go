package appview

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// PDSDiscoveryHandler handles PDS discovery endpoints
func (h *Handlers) PDSDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	// Determine which discovery endpoint based on the URL path
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/discover"):
		h.handlePDSDiscover(w, r)
	case strings.HasSuffix(path, "/info"):
		h.handlePDSInfo(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handlePDSDiscover handles GET /api/v1/pds/discover?identifier={handle_or_did}
func (h *Handlers) handlePDSDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	identifier := r.URL.Query().Get("identifier")
	if identifier == "" {
		http.Error(w, "identifier parameter required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Discovering PDS for identifier", "identifier", identifier)

	// Resolve identifier to PDS endpoint
	pdsInfo, err := h.discoverPDSEndpoint(r.Context(), identifier)
	if err != nil {
		h.logger.Error("PDS discovery failed", "error", err, "identifier", identifier)
		http.Error(w, "PDS discovery failed", http.StatusNotFound)
		return
	}

	// Return PDS information
	response := map[string]interface{}{
		"identifier":             identifier,
		"pds_endpoint":           pdsInfo.Endpoint,
		"pds_name":               pdsInfo.Name,
		"pds_description":        pdsInfo.Description,
		"supports_oauth":         pdsInfo.SupportsOAuth,
		"supports_external_auth": pdsInfo.SupportsExternalAuth,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handlePDSInfo handles GET /api/v1/pds/info?endpoint={url}
func (h *Handlers) handlePDSInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		http.Error(w, "endpoint parameter required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Getting PDS info", "endpoint", endpoint)

	// Get PDS server information
	pdsInfo, err := h.getPDSServerInfo(r.Context(), endpoint)
	if err != nil {
		h.logger.Error("Failed to get PDS info", "error", err, "endpoint", endpoint)
		http.Error(w, "Failed to get PDS info", http.StatusBadRequest)
		return
	}

	// Return PDS server information
	response := map[string]interface{}{
		"endpoint":               endpoint,
		"name":                   pdsInfo.Name,
		"description":            pdsInfo.Description,
		"version":                pdsInfo.Version,
		"supports_oauth":         pdsInfo.SupportsOAuth,
		"supports_external_auth": pdsInfo.SupportsExternalAuth,
		"public_key":             pdsInfo.PublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PDSInfo represents information about a PDS server
type PDSInfo struct {
	Endpoint             string `json:"endpoint"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Version              string `json:"version"`
	SupportsOAuth        bool   `json:"supports_oauth"`
	SupportsExternalAuth bool   `json:"supports_external_auth"`
	PublicKey            string `json:"public_key,omitempty"`
}

// discoverPDSEndpoint discovers the PDS endpoint for a given identifier
func (h *Handlers) discoverPDSEndpoint(ctx context.Context, identifier string) (*PDSInfo, error) {
	// For now, this is a simplified implementation
	// In production, this would:
	// 1. Parse the identifier (handle or DID)
	// 2. Resolve the identifier to get the DID
	// 3. Fetch the DID document
	// 4. Extract the PDS endpoint from the DID document
	// 5. Query the PDS server for its capabilities

	h.logger.Debug("Discovering PDS endpoint", "identifier", identifier)

	// Mock discovery for development
	if strings.Contains(identifier, ".hashpost.local") {
		return &PDSInfo{
			Endpoint:             "http://localhost:8080",
			Name:                 "HashPost PDS",
			Description:          "HashPost's hosted Personal Data Server",
			Version:              "1.0.0",
			SupportsOAuth:        true,
			SupportsExternalAuth: true,
		}, nil
	}

	// For external identifiers, we'd need to resolve them
	// For now, return a mock external PDS
	return &PDSInfo{
		Endpoint:             "https://external-pds.example.com",
		Name:                 "External PDS",
		Description:          "External Personal Data Server",
		Version:              "1.0.0",
		SupportsOAuth:        true,
		SupportsExternalAuth: true,
	}, nil
}

// getPDSServerInfo gets information about a PDS server
func (h *Handlers) getPDSServerInfo(ctx context.Context, endpoint string) (*PDSInfo, error) {
	// For now, this is a simplified implementation
	// In production, this would:
	// 1. Make a request to the PDS server's describe endpoint
	// 2. Parse the response to get server capabilities
	// 3. Cache the information for future use

	h.logger.Debug("Getting PDS server info", "endpoint", endpoint)

	// Mock server info for development
	return &PDSInfo{
		Endpoint:             endpoint,
		Name:                 "Mock PDS Server",
		Description:          "Mock Personal Data Server for development",
		Version:              "1.0.0",
		SupportsOAuth:        true,
		SupportsExternalAuth: true,
		PublicKey:            "mock_public_key",
	}, nil
}
