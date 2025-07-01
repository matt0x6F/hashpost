package handlers

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
)

// CorrelationHandler handles administrative correlation requests
type CorrelationHandler struct {
	db                 bob.Executor
	ibeSystem          *ibe.IBESystem
	securePseudonymDAO *dao.SecurePseudonymDAO
	identityMappingDAO *dao.IdentityMappingDAO
	postDAO            *dao.PostDAO
	commentDAO         *dao.CommentDAO
	subforumDAO        *dao.SubforumDAO
}

// NewCorrelationHandler creates a new correlation handler
func NewCorrelationHandler(db bob.Executor, ibeSystem *ibe.IBESystem, securePseudonymDAO *dao.SecurePseudonymDAO, identityMappingDAO *dao.IdentityMappingDAO, postDAO *dao.PostDAO, commentDAO *dao.CommentDAO, subforumDAO *dao.SubforumDAO) *CorrelationHandler {
	return &CorrelationHandler{
		db:                 db,
		ibeSystem:          ibeSystem,
		securePseudonymDAO: securePseudonymDAO,
		identityMappingDAO: identityMappingDAO,
		postDAO:            postDAO,
		commentDAO:         commentDAO,
		subforumDAO:        subforumDAO,
	}
}

// RequestFingerprintCorrelation handles fingerprint-based correlation for moderation
func (h *CorrelationHandler) RequestFingerprintCorrelation(ctx context.Context, input *models.FingerprintCorrelationInput) (*models.FingerprintCorrelationResponse, error) {
	// Extract admin from context (from admin JWT token)
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	adminID := userCtx.UserID

	log.Info().
		Str("endpoint", "admin/correlation/fingerprint").
		Str("component", "handler").
		Int64("admin_id", adminID).
		Str("requested_pseudonym", input.Body.RequestedPseudonym).
		Str("justification", input.Body.Justification).
		Int("subforum_id", input.Body.SubforumID).
		Str("incident_id", input.Body.IncidentID).
		Msg("Fingerprint correlation requested")

	// Validate admin permissions
	if !userCtx.HasCapability("correlate_fingerprints") {
		log.Warn().
			Int64("admin_id", adminID).
			Msg("User lacks correlate_fingerprints capability")
		return nil, fmt.Errorf("insufficient permissions: correlate_fingerprints capability required")
	}

	// Check if pseudonym exists
	pseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, input.Body.RequestedPseudonym)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", input.Body.RequestedPseudonym).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		log.Warn().
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}

	// Generate correlation ID
	correlationID := uuid.Must(uuid.NewV4()).String()

	// Perform IBE correlation
	// Get the identity mapping for the requested pseudonym
	identityMapping, err := h.identityMappingDAO.GetIdentityMappingByPseudonymID(ctx, input.Body.RequestedPseudonym)
	if err != nil {
		log.Error().Err(err).
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Failed to get identity mapping")
		return nil, fmt.Errorf("failed to get identity mapping: %w", err)
	}
	if identityMapping == nil {
		log.Warn().
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Identity mapping not found")
		return nil, fmt.Errorf("identity mapping not found for pseudonym")
	}

	// Generate admin key for decryption based on user's role
	adminKey := h.ibeSystem.GenerateRoleKey("moderator", "subforum_correlation", time.Now().AddDate(0, 1, 0))

	// Decrypt the identity mapping to get the fingerprint
	decryptedMapping, _, err := h.ibeSystem.DecryptIdentity(identityMapping.EncryptedRealIdentity, adminKey)
	if err != nil {
		log.Error().Err(err).
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Failed to decrypt identity mapping")
		return nil, fmt.Errorf("failed to decrypt identity mapping: %w", err)
	}

	// Parse the decrypted mapping to extract fingerprint
	// Format should be "fingerprint:pseudonymID"
	mappingParts := strings.Split(decryptedMapping, ":")
	if len(mappingParts) != 2 {
		log.Error().
			Str("decrypted_mapping", decryptedMapping).
			Msg("Invalid decrypted mapping format")
		return nil, fmt.Errorf("invalid decrypted mapping format")
	}
	fingerprint := mappingParts[0]

	// Find all pseudonyms that share the same fingerprint
	relatedMappings, err := h.identityMappingDAO.GetIdentityMappingsByFingerprint(ctx, fingerprint)
	if err != nil {
		log.Error().Err(err).
			Str("fingerprint", fingerprint).
			Msg("Failed to get related identity mappings")
		return nil, fmt.Errorf("failed to get related identity mappings: %w", err)
	}

	// Build correlation results
	results := make([]models.CorrelationResult, 0, len(relatedMappings))
	for _, mapping := range relatedMappings {
		// Get pseudonym details
		pseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, mapping.PseudonymID)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", mapping.PseudonymID).Msg("Failed to get pseudonym from database")
			continue
		}
		if pseudonym == nil {
			log.Warn().
				Str("pseudonym_id", mapping.PseudonymID).
				Msg("Pseudonym not found")
			continue
		}

		// Get actual post/comment counts for the specific subforum
		subforumID := int32(input.Body.SubforumID)
		postsInSubforum, err := h.postDAO.CountPostsByPseudonymInSubforum(ctx, mapping.PseudonymID, subforumID)
		if err != nil {
			log.Error().Err(err).
				Str("pseudonym_id", mapping.PseudonymID).
				Int32("subforum_id", subforumID).
				Msg("Failed to count posts in subforum")
			postsInSubforum = 0
		}

		commentsInSubforum, err := h.commentDAO.CountCommentsByPseudonymInSubforum(ctx, mapping.PseudonymID, subforumID)
		if err != nil {
			log.Error().Err(err).
				Str("pseudonym_id", mapping.PseudonymID).
				Int32("subforum_id", subforumID).
				Msg("Failed to count comments in subforum")
			commentsInSubforum = 0
		}

		postsInSubforumInt := int(postsInSubforum)
		commentsInSubforumInt := int(commentsInSubforum)

		result := models.CorrelationResult{
			PseudonymID:        mapping.PseudonymID,
			DisplayName:        pseudonym.DisplayName,
			CreatedAt:          pseudonym.CreatedAt.V.Format(time.RFC3339),
			PostsInSubforum:    &postsInSubforumInt,
			CommentsInSubforum: &commentsInSubforumInt,
		}
		results = append(results, result)
	}

	// Create audit record
	auditID := uuid.Must(uuid.NewV4())

	// Serialize correlation results for audit
	correlationResultJSON, err := json.Marshal(results)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal correlation results")
		return nil, fmt.Errorf("failed to serialize correlation results: %w", err)
	}

	// Create SQL null types
	requestedFingerprint := sql.Null[string]{}
	if input.Body.RequestedFingerprint != "" {
		requestedFingerprint.Scan(input.Body.RequestedFingerprint)
	}

	correlationResult := sql.Null[types.JSON[json.RawMessage]]{}
	correlationResult.Scan(correlationResultJSON)

	timestamp := sql.Null[time.Time]{}
	timestamp.Scan(time.Now())

	incidentID := sql.Null[string]{}
	if input.Body.IncidentID != "" {
		incidentID.Scan(input.Body.IncidentID)
	}

	requestSource := sql.Null[string]{}
	requestSource.Scan("manual")

	// Create correlation audit record
	auditRecord := &dbmodels.CorrelationAuditSetter{
		AuditID:              &auditID,
		UserID:               &adminID,
		PseudonymID:          &pseudonym.PseudonymID,
		AdminUsername:        &userCtx.Email,
		RoleUsed:             &[]string{"moderator"}[0],
		RequestedPseudonym:   &input.Body.RequestedPseudonym,
		RequestedFingerprint: &requestedFingerprint,
		Justification:        &input.Body.Justification,
		CorrelationType:      &[]string{"fingerprint"}[0],
		CorrelationResult:    &correlationResult,
		Timestamp:            &timestamp,
		IncidentID:           &incidentID,
		RequestSource:        &requestSource,
	}

	// Store audit record in database
	_, err = dbmodels.CorrelationAudits.Insert(auditRecord).One(ctx, h.db)
	if err != nil {
		log.Error().Err(err).
			Str("audit_id", auditID.String()).
			Msg("Failed to create correlation audit record")
		return nil, fmt.Errorf("failed to create audit record: %w", err)
	}

	response := models.NewFingerprintCorrelationResponse(correlationID, results, auditID.String())

	log.Info().
		Str("endpoint", "admin/correlation/fingerprint").
		Str("component", "handler").
		Int64("admin_id", adminID).
		Str("correlation_id", correlationID).
		Int("results_count", len(results)).
		Str("audit_id", auditID.String()).
		Msg("Fingerprint correlation completed")

	return response, nil
}

// RequestIdentityCorrelation handles identity-based correlation for platform-wide investigations
func (h *CorrelationHandler) RequestIdentityCorrelation(ctx context.Context, input *models.IdentityCorrelationInput) (*models.IdentityCorrelationResponse, error) {
	// Extract admin from context (from admin JWT token)
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	adminID := userCtx.UserID

	log.Info().
		Str("endpoint", "admin/correlation/identity").
		Str("component", "handler").
		Int64("admin_id", adminID).
		Str("requested_pseudonym", input.Body.RequestedPseudonym).
		Str("justification", input.Body.Justification).
		Str("legal_basis", input.Body.LegalBasis).
		Str("incident_id", input.Body.IncidentID).
		Str("scope", input.Body.Scope).
		Msg("Identity correlation requested")

	// Validate admin permissions for platform-wide correlation
	if !userCtx.HasCapability("correlate_identities") {
		log.Warn().
			Int64("admin_id", adminID).
			Msg("User lacks correlate_identities capability")
		return nil, fmt.Errorf("insufficient permissions: correlate_identities capability required")
	}

	// Check if pseudonym exists
	pseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, input.Body.RequestedPseudonym)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", input.Body.RequestedPseudonym).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		log.Warn().
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}

	// Generate correlation ID
	correlationID := uuid.Must(uuid.NewV4()).String()

	// Perform IBE identity correlation
	// Get the identity mapping for the requested pseudonym
	identityMapping, err := h.identityMappingDAO.GetIdentityMappingByPseudonymID(ctx, input.Body.RequestedPseudonym)
	if err != nil {
		log.Error().Err(err).
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Failed to get identity mapping")
		return nil, fmt.Errorf("failed to get identity mapping: %w", err)
	}
	if identityMapping == nil {
		log.Warn().
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Identity mapping not found")
		return nil, fmt.Errorf("identity mapping not found for pseudonym")
	}

	// Generate admin key for decryption based on user's role
	adminKey := h.ibeSystem.GenerateRoleKey("site_admin", "full_correlation", time.Now().AddDate(0, 1, 0))

	// Decrypt the identity mapping to get the fingerprint
	decryptedMapping, _, err := h.ibeSystem.DecryptIdentity(identityMapping.EncryptedRealIdentity, adminKey)
	if err != nil {
		log.Error().Err(err).
			Str("requested_pseudonym", input.Body.RequestedPseudonym).
			Msg("Failed to decrypt identity mapping")
		return nil, fmt.Errorf("failed to decrypt identity mapping: %w", err)
	}

	// Parse the decrypted mapping to extract fingerprint
	// Format should be "fingerprint:pseudonymID"
	mappingParts := strings.Split(decryptedMapping, ":")
	if len(mappingParts) != 2 {
		log.Error().
			Str("decrypted_mapping", decryptedMapping).
			Msg("Invalid decrypted mapping format")
		return nil, fmt.Errorf("invalid decrypted mapping format")
	}
	fingerprint := mappingParts[0]

	// Find all pseudonyms that share the same fingerprint (platform-wide correlation)
	relatedMappings, err := h.identityMappingDAO.GetIdentityMappingsByFingerprint(ctx, fingerprint)
	if err != nil {
		log.Error().Err(err).
			Str("fingerprint", fingerprint).
			Msg("Failed to get related identity mappings")
		return nil, fmt.Errorf("failed to get related identity mappings: %w", err)
	}

	// Build correlation results
	results := make([]models.CorrelationResult, 0, len(relatedMappings))
	for _, mapping := range relatedMappings {
		// Get pseudonym details
		pseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, mapping.PseudonymID)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", mapping.PseudonymID).Msg("Failed to get pseudonym from database")
			continue
		}
		if pseudonym == nil {
			log.Warn().
				Str("pseudonym_id", mapping.PseudonymID).
				Msg("Pseudonym not found")
			continue
		}

		// Get actual post/comment counts and subforum activity
		totalPosts, err := h.postDAO.CountPostsByPseudonym(ctx, mapping.PseudonymID)
		if err != nil {
			log.Error().Err(err).
				Str("pseudonym_id", mapping.PseudonymID).
				Msg("Failed to count total posts")
			totalPosts = 0
		}

		totalComments, err := h.commentDAO.CountCommentsByPseudonym(ctx, mapping.PseudonymID)
		if err != nil {
			log.Error().Err(err).
				Str("pseudonym_id", mapping.PseudonymID).
				Msg("Failed to count total comments")
			totalComments = 0
		}

		// Get subforums where the pseudonym has been active
		postSubforums, err := h.postDAO.GetSubforumsByPseudonym(ctx, mapping.PseudonymID)
		if err != nil {
			log.Error().Err(err).
				Str("pseudonym_id", mapping.PseudonymID).
				Msg("Failed to get subforums by posts")
			postSubforums = []int32{}
		}

		commentSubforums, err := h.commentDAO.GetSubforumsByPseudonymComments(ctx, mapping.PseudonymID)
		if err != nil {
			log.Error().Err(err).
				Str("pseudonym_id", mapping.PseudonymID).
				Msg("Failed to get subforums by comments")
			commentSubforums = []int32{}
		}

		// Combine and deduplicate subforums
		subforumMap := make(map[int32]bool)
		for _, sf := range postSubforums {
			subforumMap[sf] = true
		}
		for _, sf := range commentSubforums {
			subforumMap[sf] = true
		}

		// Convert to string slice with actual subforum names from the database
		subforumsActive := make([]string, 0, len(subforumMap))
		for subforumID := range subforumMap {
			// Get subforum details from database
			subforum, err := h.subforumDAO.GetSubforumByID(ctx, subforumID)
			if err != nil {
				log.Error().Err(err).
					Int32("subforum_id", subforumID).
					Msg("Failed to get subforum details")
				// Fallback to ID if name lookup fails
				subforumsActive = append(subforumsActive, fmt.Sprintf("subforum_%d", subforumID))
				continue
			}
			if subforum == nil {
				log.Warn().
					Int32("subforum_id", subforumID).
					Msg("Subforum not found")
				// Fallback to ID if subforum doesn't exist
				subforumsActive = append(subforumsActive, fmt.Sprintf("subforum_%d", subforumID))
				continue
			}
			// Use the subforum name (not display name) for consistency
			subforumsActive = append(subforumsActive, subforum.Name)
		}

		totalPostsInt := int(totalPosts)
		totalCommentsInt := int(totalComments)

		// For identity correlation, we include the encrypted real identity
		encryptedIdentity := hex.EncodeToString(mapping.EncryptedRealIdentity)

		result := models.CorrelationResult{
			PseudonymID:           mapping.PseudonymID,
			DisplayName:           pseudonym.DisplayName,
			EncryptedRealIdentity: &encryptedIdentity,
			CreatedAt:             pseudonym.CreatedAt.V.Format(time.RFC3339),
			TotalPosts:            &totalPostsInt,
			TotalComments:         &totalCommentsInt,
			SubforumsActive:       &subforumsActive,
		}
		results = append(results, result)
	}

	// Create audit record
	auditID := uuid.Must(uuid.NewV4())

	// Serialize correlation results for audit
	correlationResultJSON, err := json.Marshal(results)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal correlation results")
		return nil, fmt.Errorf("failed to serialize correlation results: %w", err)
	}

	// Create SQL null types
	requestedFingerprint := sql.Null[string]{}
	if input.Body.RequestedFingerprint != "" {
		requestedFingerprint.Scan(input.Body.RequestedFingerprint)
	}

	correlationResult := sql.Null[types.JSON[json.RawMessage]]{}
	correlationResult.Scan(correlationResultJSON)

	timestamp := sql.Null[time.Time]{}
	timestamp.Scan(time.Now())

	legalBasis := sql.Null[string]{}
	if input.Body.LegalBasis != "" {
		legalBasis.Scan(input.Body.LegalBasis)
	}

	incidentID := sql.Null[string]{}
	if input.Body.IncidentID != "" {
		incidentID.Scan(input.Body.IncidentID)
	}

	requestSource := sql.Null[string]{}
	requestSource.Scan("manual")

	// Create correlation audit record
	auditRecord := &dbmodels.CorrelationAuditSetter{
		AuditID:              &auditID,
		UserID:               &adminID,
		PseudonymID:          &pseudonym.PseudonymID,
		AdminUsername:        &userCtx.Email,
		RoleUsed:             &[]string{"site_admin"}[0],
		RequestedPseudonym:   &input.Body.RequestedPseudonym,
		RequestedFingerprint: &requestedFingerprint,
		Justification:        &input.Body.Justification,
		CorrelationType:      &[]string{"identity"}[0],
		CorrelationResult:    &correlationResult,
		Timestamp:            &timestamp,
		LegalBasis:           &legalBasis,
		IncidentID:           &incidentID,
		RequestSource:        &requestSource,
	}

	// Store audit record in database
	_, err = dbmodels.CorrelationAudits.Insert(auditRecord).One(ctx, h.db)
	if err != nil {
		log.Error().Err(err).
			Str("audit_id", auditID.String()).
			Msg("Failed to create correlation audit record")
		return nil, fmt.Errorf("failed to create audit record: %w", err)
	}

	response := models.NewIdentityCorrelationResponse(correlationID, results, auditID.String())

	log.Info().
		Str("endpoint", "admin/correlation/identity").
		Str("component", "handler").
		Int64("admin_id", adminID).
		Str("correlation_id", correlationID).
		Int("results_count", len(results)).
		Str("audit_id", auditID.String()).
		Msg("Identity correlation completed")

	return response, nil
}

// GetCorrelationHistory handles getting correlation request history
func (h *CorrelationHandler) GetCorrelationHistory(ctx context.Context, input *models.CorrelationHistoryInput) (*models.CorrelationHistoryResponse, error) {
	// Extract admin from context (from admin JWT token)
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	adminID := userCtx.UserID

	log.Info().
		Str("endpoint", "admin/correlation/history").
		Str("component", "handler").
		Int64("admin_id", adminID).
		Str("correlation_type", input.CorrelationType).
		Msg("Get correlation history requested")

	// Validate admin permissions
	if !userCtx.HasCapability("view_correlation_history") {
		log.Warn().
			Int64("admin_id", adminID).
			Msg("User lacks view_correlation_history capability")
		return nil, fmt.Errorf("insufficient permissions: view_correlation_history capability required")
	}

	// Get correlation history from database with basic pagination
	// Note: The CorrelationAudits table uses a ViewQuery type which has different methods
	// than regular SelectQuery. For now, we'll use a simple approach.
	auditRecords, err := dbmodels.CorrelationAudits.Query().All(ctx, h.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get correlation history")
		return nil, fmt.Errorf("failed to get correlation history: %w", err)
	}

	// Apply correlation type filter if specified
	if input.CorrelationType != "" {
		filteredRecords := make(dbmodels.CorrelationAuditSlice, 0)
		for _, record := range auditRecords {
			if record.CorrelationType == input.CorrelationType {
				filteredRecords = append(filteredRecords, record)
			}
		}
		auditRecords = filteredRecords
	}

	// Apply pagination
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100 // Cap at 100 records per page
	}

	offset := (input.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	total := len(auditRecords)

	// Apply pagination manually
	if offset >= len(auditRecords) {
		auditRecords = dbmodels.CorrelationAuditSlice{}
	} else {
		end := offset + limit
		if end > len(auditRecords) {
			end = len(auditRecords)
		}
		auditRecords = auditRecords[offset:end]
	}

	// Convert audit records to correlation models
	correlations := make([]models.Correlation, 0, len(auditRecords))
	for _, record := range auditRecords {
		// Parse correlation result to get results count
		resultsCount := 0
		if record.CorrelationResult.Valid {
			// Convert types.JSON to []byte for unmarshaling
			jsonBytes, err := record.CorrelationResult.V.MarshalJSON()
			if err == nil {
				var results []models.CorrelationResult
				if err := json.Unmarshal(jsonBytes, &results); err == nil {
					resultsCount = len(results)
				}
			}
		}

		correlation := models.Correlation{
			CorrelationID:      record.AuditID.String(),
			CorrelationType:    record.CorrelationType,
			RequestedPseudonym: record.RequestedPseudonym,
			Justification:      record.Justification,
			Status:             "completed", // All stored records are completed
			Timestamp:          record.Timestamp.V.Format(time.RFC3339),
			ResultsCount:       resultsCount,
		}
		correlations = append(correlations, correlation)
	}

	response := models.NewCorrelationHistoryResponse(correlations, input.Page, input.Limit, int(total))

	log.Info().
		Str("endpoint", "admin/correlation/history").
		Str("component", "handler").
		Int64("admin_id", adminID).
		Int("count", len(correlations)).
		Int("total", int(total)).
		Msg("Get correlation history completed")

	return response, nil
}
