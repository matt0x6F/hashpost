package pds

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/matt0x6f/hashpost/internal/lexicons"
)

// handleCreateRecord handles record creation
func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := authHeader[7:]
	session, authErr := s.authService.ValidateToken(token)
	if authErr != nil {
		s.logger.Error("Token validation failed", "error", authErr)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	s.logger.Debug("Authenticated user", "did", session.DID, "handle", session.Handle)

	// Parse request body
	var req struct {
		Repo       string                 `json:"repo"`
		Collection string                 `json:"collection"`
		Record     map[string]interface{} `json:"record"`
		Validate   bool                   `json:"validate,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Repo == "" || req.Collection == "" || req.Record == nil {
		http.Error(w, "repo, collection, and record required", http.StatusBadRequest)
		return
	}

	// Validate collection
	validCollections := []string{
		lexicons.CollectionFeedPost,
		lexicons.CollectionFeedSubforum,
		"app.bsky.feed.post",
		"app.bsky.feed.like",
		"app.bsky.feed.repost",
		"app.bsky.graph.follow",
		"app.bsky.actor.profile",
	}

	isValidCollection := false
	for _, valid := range validCollections {
		if req.Collection == valid {
			isValidCollection = true
			break
		}
	}

	if !isValidCollection {
		http.Error(w, "Invalid collection: "+req.Collection, http.StatusBadRequest)
		return
	}

	s.logger.Debug("Creating record", "repo", req.Repo, "collection", req.Collection)

	// Generate record ID and URI
	recordID := uuid.New().String()
	uri := fmt.Sprintf("at://%s/%s/%s", req.Repo, req.Collection, recordID)

	// Compute proper content-addressed CID for the record
	computedCID, err := s.cidService.ComputeRecordCID(r.Context(), req.Record)
	if err != nil {
		s.logger.Error("Failed to compute CID for record", "error", err)
		http.Error(w, "Failed to compute record CID", http.StatusInternalServerError)
		return
	}
	cid := computedCID

	// Store record in database based on collection type
	ctx := r.Context()

	switch req.Collection {
	case lexicons.CollectionFeedPost:
		err = s.createHashPostRecord(ctx, req.Repo, req.Collection, recordID, uri, cid, req.Record)
	case lexicons.CollectionFeedSubforum:
		err = s.createHashPostRecord(ctx, req.Repo, req.Collection, recordID, uri, cid, req.Record)
	default:
		// For other collections, store as generic record
		err = s.createGenericRecord(ctx, req.Repo, req.Collection, recordID, uri, cid, req.Record)
	}

	if err != nil {
		s.logger.Error("Failed to create record", "error", err, "collection", req.Collection)
		http.Error(w, "Failed to create record", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"uri":    uri,
		"cid":    cid,
		"record": req.Record,
	}

	// Publish record created event
	if err := s.eventStream.PublishRecordEvent(ctx, EventTypeRecordCreated, req.Repo, req.Collection, req.Record, uri, cid); err != nil {
		s.logger.Error("Failed to publish record created event", "error", err)
		// Don't fail the request, just log the error
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetRecord handles record retrieval
func (s *Server) handleGetRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Check authentication
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authorization header required"})
		return
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid authorization header format"})
		return
	}

	token := authHeader[7:]
	session, authErr := s.authService.ValidateToken(token)
	if authErr != nil {
		s.logger.Error("Token validation failed", "error", authErr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
		return
	}

	s.logger.Debug("Authenticated user", "did", session.DID, "handle", session.Handle)

	repo := r.URL.Query().Get("repo")
	collection := r.URL.Query().Get("collection")
	rkey := r.URL.Query().Get("rkey")

	if repo == "" || collection == "" || rkey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "repo, collection, and rkey required"})
		return
	}

	s.logger.Debug("Getting record", "repo", repo, "collection", collection, "rkey", rkey)

	ctx := r.Context()
	uri := fmt.Sprintf("at://%s/%s/%s", repo, collection, rkey)

	// Look up record in database based on collection type
	var response map[string]interface{}
	var err error

	switch collection {
	case lexicons.CollectionFeedPost:
		response, err = s.getPostRecord(ctx, uri)
	case lexicons.CollectionFeedSubforum:
		response, err = s.getSubforumRecord(ctx, uri)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unsupported collection"})
		return
	}

	if err != nil {
		s.logger.Error("Failed to get record", "error", err, "uri", uri)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Record not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleListRecords handles record listing
func (s *Server) handleListRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := authHeader[7:]
	session, authErr := s.authService.ValidateToken(token)
	if authErr != nil {
		s.logger.Error("Token validation failed", "error", authErr)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	s.logger.Debug("Authenticated user", "did", session.DID, "handle", session.Handle)

	repo := r.URL.Query().Get("repo")
	collection := r.URL.Query().Get("collection")
	limit := r.URL.Query().Get("limit")
	cursor := r.URL.Query().Get("cursor")

	if repo == "" || collection == "" {
		http.Error(w, "repo and collection required", http.StatusBadRequest)
		return
	}

	s.logger.Debug("Listing records", "repo", repo, "collection", collection, "limit", limit, "cursor", cursor)

	ctx := r.Context()

	// Parse limit with default
	limitInt := 50
	if limit != "" {
		if parsed, err := fmt.Sscanf(limit, "%d", &limitInt); err != nil || parsed != 1 {
			limitInt = 50
		}
	}

	// Query records from database based on collection type
	var response map[string]interface{}
	var err error

	switch collection {
	case lexicons.CollectionFeedPost:
		response, err = s.listPostRecords(ctx, repo, limitInt, cursor)
	case lexicons.CollectionFeedSubforum:
		response, err = s.listSubforumRecords(ctx, repo, limitInt, cursor)
	default:
		http.Error(w, "Unsupported collection", http.StatusBadRequest)
		return
	}

	if err != nil {
		s.logger.Error("Failed to list records", "error", err, "collection", collection)
		http.Error(w, "Failed to list records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handlePutRecord handles record updates
func (s *Server) handlePutRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := authHeader[7:]
	session, authErr := s.authService.ValidateToken(token)
	if authErr != nil {
		s.logger.Error("Token validation failed", "error", authErr)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	s.logger.Debug("Authenticated user", "did", session.DID, "handle", session.Handle)

	// Parse request body
	var req struct {
		Repo       string                 `json:"repo"`
		Collection string                 `json:"collection"`
		Rkey       string                 `json:"rkey"`
		Record     map[string]interface{} `json:"record"`
		Validate   bool                   `json:"validate,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Repo == "" || req.Collection == "" || req.Rkey == "" || req.Record == nil {
		http.Error(w, "repo, collection, rkey, and record required", http.StatusBadRequest)
		return
	}

	s.logger.Debug("Updating record", "repo", req.Repo, "collection", req.Collection, "rkey", req.Rkey)

	ctx := r.Context()
	uri := fmt.Sprintf("at://%s/%s/%s", req.Repo, req.Collection, req.Rkey)

	// Update record in database based on collection type
	var response map[string]interface{}
	var err error

	switch req.Collection {
	case lexicons.CollectionFeedPost:
		response, err = s.updatePostRecord(ctx, uri, req.Record)
	case lexicons.CollectionFeedSubforum:
		response, err = s.updateSubforumRecord(ctx, req.Rkey, req.Record)
	default:
		http.Error(w, "Unsupported collection", http.StatusBadRequest)
		return
	}

	if err != nil {
		s.logger.Error("Failed to update record", "error", err, "uri", uri)
		http.Error(w, "Failed to update record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleDeleteRecord handles record deletion
func (s *Server) handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := authHeader[7:]
	session, authErr := s.authService.ValidateToken(token)
	if authErr != nil {
		s.logger.Error("Token validation failed", "error", authErr)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	s.logger.Debug("Authenticated user", "did", session.DID, "handle", session.Handle)

	// Parse request body
	var req struct {
		Repo       string `json:"repo"`
		Collection string `json:"collection"`
		Rkey       string `json:"rkey"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Repo == "" || req.Collection == "" || req.Rkey == "" {
		http.Error(w, "repo, collection, and rkey required", http.StatusBadRequest)
		return
	}

	s.logger.Debug("Deleting record", "repo", req.Repo, "collection", req.Collection, "rkey", req.Rkey)

	ctx := r.Context()
	uri := fmt.Sprintf("at://%s/%s/%s", req.Repo, req.Collection, req.Rkey)

	// Delete record from database based on collection type
	var err error

	switch req.Collection {
	case lexicons.CollectionFeedPost:
		err = s.deletePostRecord(ctx, uri)
	case lexicons.CollectionFeedSubforum:
		err = s.deleteSubforumRecord(ctx, req.Rkey)
	default:
		http.Error(w, "Unsupported collection", http.StatusBadRequest)
		return
	}

	if err != nil {
		s.logger.Error("Failed to delete record", "error", err, "uri", uri)
		http.Error(w, "Failed to delete record", http.StatusInternalServerError)
		return
	}

	s.logger.Info("Record deleted", "uri", uri)
	w.WriteHeader(http.StatusOK)
}
