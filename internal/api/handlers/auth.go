package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/api/validation"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	config             *config.Config
	db                 bob.Executor
	userDAO            *dao.UserDAO
	securePseudonymDAO *dao.SecurePseudonymDAO
	identityMappingDAO *dao.IdentityMappingDAO
	ibeSystem          *ibe.IBESystem
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(cfg *config.Config, db bob.Executor, rawDB *sql.DB) *AuthHandler {
	userDAO := dao.NewUserDAO(db)
	ibeSystem := ibe.NewIBESystem()
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	securePseudonymDAO := dao.NewSecurePseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)

	return &AuthHandler{
		config:             cfg,
		db:                 db,
		userDAO:            userDAO,
		securePseudonymDAO: securePseudonymDAO,
		identityMappingDAO: identityMappingDAO,
		ibeSystem:          ibeSystem,
	}
}

// NewAuthHandlerWithIBE creates a new authentication handler with a specific IBE system
func NewAuthHandlerWithIBE(cfg *config.Config, db bob.Executor, rawDB *sql.DB, ibeSystem *ibe.IBESystem) *AuthHandler {
	userDAO := dao.NewUserDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	securePseudonymDAO := dao.NewSecurePseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)

	return &AuthHandler{
		config:             cfg,
		db:                 db,
		userDAO:            userDAO,
		securePseudonymDAO: securePseudonymDAO,
		identityMappingDAO: identityMappingDAO,
		ibeSystem:          ibeSystem,
	}
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(ctx context.Context, input *models.UserRegistrationInput) (*models.UserRegistrationResponse, error) {
	log.Info().
		Str("endpoint", "auth/register").
		Str("component", "auth_handler").
		Msg("Processing user registration request")

	// Enhanced validation using the validation package
	if err := validation.ValidateEmail(input.Body.Email); err != nil {
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}

	if err := validation.ValidatePassword(input.Body.Password, h.config.Security.PasswordValidation); err != nil {
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}

	if err := validation.ValidateDisplayName(input.Body.DisplayName); err != nil {
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}

	// Check if user already exists
	existingUser, err := h.userDAO.GetUserByEmail(ctx, input.Body.Email)
	if err == nil && existingUser != nil {
		log.Warn().
			Str("email", input.Body.Email).
			Msg("User registration failed - email already exists")
		return nil, fmt.Errorf("user with email %s already exists", input.Body.Email)
	}

	// Hash password
	hashedPassword := h.hashPassword(input.Body.Password)

	// Create user
	user, err := h.userDAO.CreateUser(ctx, input.Body.Email, hashedPassword)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", input.Body.Email).
			Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default role keys for the user
	roleKeyDAO := dao.NewRoleKeyDAO(h.db)
	if err := roleKeyDAO.EnsureDefaultKeys(ctx, h.ibeSystem, user.UserID); err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Msg("Failed to create default role keys")
		return nil, fmt.Errorf("failed to create default role keys: %w", err)
	}

	// Create pseudonym for the user
	pseudonym, err := h.securePseudonymDAO.CreatePseudonymWithIdentityMapping(ctx, user.UserID, input.Body.DisplayName)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Str("display_name", input.Body.DisplayName).
			Msg("Failed to create pseudonym in database")
		return nil, fmt.Errorf("failed to create pseudonym: %w", err)
	}

	// Get user roles and capabilities from database
	roles := []string{"user"}                                                                  // Default role
	capabilities := []string{"create_content", "vote", "message", "report", "create_subforum"} // Default capabilities

	// If user has roles/capabilities stored in database, use those
	if user.Roles.Valid {
		rawValue, err := user.Roles.V.Value()
		if err == nil {
			var userRoles []string
			if err := json.Unmarshal(rawValue.([]byte), &userRoles); err == nil && len(userRoles) > 0 {
				roles = userRoles
			}
		}
	}

	if user.Capabilities.Valid {
		rawValue, err := user.Capabilities.V.Value()
		if err == nil {
			var userCapabilities []string
			if err := json.Unmarshal(rawValue.([]byte), &userCapabilities); err == nil && len(userCapabilities) > 0 {
				capabilities = userCapabilities
			}
		}
	}

	// Create user context for JWT generation
	userCtx := &middleware.UserContext{
		UserID:            user.UserID,
		Email:             user.Email,
		Roles:             roles,
		Capabilities:      capabilities,
		MFAEnabled:        false, // TODO: Implement MFA
		ActivePseudonymID: pseudonym.PseudonymID,
		DisplayName:       pseudonym.DisplayName,
	}

	// Generate JWT tokens
	accessToken, err := middleware.GenerateJWT(userCtx, h.config.JWT.Secret, h.config.JWT.Expiration)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Msg("Failed to generate access token")
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (longer expiration)
	refreshToken, err := middleware.GenerateJWT(userCtx, h.config.JWT.Secret, 7*24*time.Hour) // 7 days
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Msg("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	log.Info().
		Int64("user_id", user.UserID).
		Str("email", input.Body.Email).
		Str("pseudonym_id", pseudonym.PseudonymID).
		Msg("User registered successfully")

	return models.NewUserRegistrationResponse(
		int(user.UserID),
		user.Email,
		roles,
		capabilities,
		pseudonym.PseudonymID,
		pseudonym.DisplayName,
		accessToken,
		refreshToken,
	), nil
}

// LoginUser handles user login
func (h *AuthHandler) LoginUser(ctx context.Context, input *models.UserLoginInput) (*models.UserLoginResponse, error) {
	log.Info().
		Str("endpoint", "auth/login").
		Str("component", "auth_handler").
		Msg("Processing user login request")

	// Enhanced validation using the validation package
	if err := validation.ValidateEmail(input.Body.Email); err != nil {
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}

	if input.Body.Password == "" {
		return nil, huma.Error422UnprocessableEntity("password is required")
	}

	// Debug: Log the input to see what we're receiving
	log.Debug().
		Str("input_email", input.Body.Email).
		Str("input_password_length", fmt.Sprintf("%d", len(input.Body.Password))).
		Msg("Login input received")

	// Find the user by email
	user, err := h.userDAO.GetUserByEmail(ctx, input.Body.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", input.Body.Email).
			Msg("Failed to find user by email")
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	if user == nil {
		log.Warn().
			Str("email", input.Body.Email).
			Msg("User not found")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive.Valid || !user.IsActive.V {
		log.Warn().
			Int64("user_id", user.UserID).
			Msg("User account is inactive")
		return nil, fmt.Errorf("account inactive")
	}

	// Check if user is suspended
	if user.IsSuspended.Valid && user.IsSuspended.V {
		log.Warn().
			Int64("user_id", user.UserID).
			Msg("User account is suspended")
		return nil, fmt.Errorf("account suspended")
	}

	// Verify password (in a real app, you'd use bcrypt.CompareHashAndPassword)
	if !h.verifyPassword(input.Body.Password, user.PasswordHash) {
		log.Warn().
			Int64("user_id", user.UserID).
			Msg("Invalid password")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last active timestamp
	err = h.userDAO.UpdateLastActive(ctx, user.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Msg("Failed to update last active timestamp")
		// Don't fail the login for this error
	}

	// Ensure default role keys exist for the user
	roleKeyDAO := dao.NewRoleKeyDAO(h.db)
	if err := roleKeyDAO.EnsureDefaultKeys(ctx, h.ibeSystem, user.UserID); err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Msg("Failed to ensure default role keys")
		return nil, fmt.Errorf("failed to ensure default role keys: %w", err)
	}

	// Get user roles and capabilities from database
	roles := []string{"user"}                                                                  // Default role
	capabilities := []string{"create_content", "vote", "message", "report", "create_subforum"} // Default capabilities

	// If user has roles/capabilities stored in database, use those
	if user.Roles.Valid {
		rawValue, err := user.Roles.V.Value()
		if err == nil {
			var userRoles []string
			if err := json.Unmarshal(rawValue.([]byte), &userRoles); err == nil && len(userRoles) > 0 {
				roles = userRoles
			}
		}
	}

	if user.Capabilities.Valid {
		rawValue, err := user.Capabilities.V.Value()
		if err == nil {
			var userCapabilities []string
			if err := json.Unmarshal(rawValue.([]byte), &userCapabilities); err == nil && len(userCapabilities) > 0 {
				capabilities = userCapabilities
			}
		}
	}

	// Get user's pseudonyms for the response
	// Use IBE-based correlation to get user's pseudonyms
	// Use the user's actual roles, not hardcoded "user"
	primaryRole := roles[0] // Use the first role for authentication
	pseudonyms, err := h.securePseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, primaryRole, "authentication")
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Str("role", primaryRole).
			Msg("Failed to get user pseudonyms")
		return nil, fmt.Errorf("failed to get user pseudonyms: %w", err)
	}

	// If user has no pseudonyms, this is a data error
	if len(pseudonyms) == 0 {
		log.Error().
			Int64("user_id", user.UserID).
			Msg("User has no pseudonyms; cannot proceed with login")
		return nil, fmt.Errorf("user has no pseudonyms; please contact support")
	}

	// Get the default pseudonym for the user
	defaultPseudonym, err := h.securePseudonymDAO.GetDefaultPseudonymByUserID(ctx, user.UserID, primaryRole, "authentication")
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Str("role", primaryRole).
			Msg("Failed to get default pseudonym")
		return nil, fmt.Errorf("failed to get default pseudonym: %w", err)
	}

	// Convert to API models
	pseudonymInfos := make([]models.PseudonymInfo, len(pseudonyms))
	for i, p := range pseudonyms {
		karmaScore := 0
		if p.KarmaScore.Valid {
			karmaScore = int(p.KarmaScore.V)
		}

		createdAt := time.Now().Format(time.RFC3339)
		if p.CreatedAt.Valid {
			createdAt = p.CreatedAt.V.Format(time.RFC3339)
		}

		lastActiveAt := time.Now().Format(time.RFC3339)
		if p.LastActiveAt.Valid {
			lastActiveAt = p.LastActiveAt.V.Format(time.RFC3339)
		}

		isActive := true
		if p.IsActive.Valid {
			isActive = p.IsActive.V
		}

		pseudonymInfos[i] = models.PseudonymInfo{
			PseudonymID:  p.PseudonymID,
			DisplayName:  p.DisplayName,
			KarmaScore:   karmaScore,
			CreatedAt:    createdAt,
			LastActiveAt: lastActiveAt,
			IsActive:     isActive,
		}
	}

	// Use the default pseudonym as the active one
	activePseudonymID := defaultPseudonym.PseudonymID
	displayName := defaultPseudonym.DisplayName

	log.Info().
		Str("active_pseudonym_id", activePseudonymID).
		Str("display_name", displayName).
		Bool("is_default", defaultPseudonym.IsDefault).
		Msg("Using default pseudonym as active pseudonym")

	// Create user context for JWT generation
	userCtx := &middleware.UserContext{
		UserID:            user.UserID,
		Email:             user.Email,
		Roles:             roles,
		Capabilities:      capabilities,
		MFAEnabled:        false, // TODO: Implement MFA
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
	}

	// Generate JWT tokens
	accessToken, err := middleware.GenerateJWT(userCtx, h.config.JWT.Secret, h.config.JWT.Expiration)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Msg("Failed to generate access token")
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (longer expiration)
	refreshToken, err := middleware.GenerateJWT(userCtx, h.config.JWT.Secret, 7*24*time.Hour) // 7 days
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", user.UserID).
			Msg("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// JWT cookies are automatically set by Huma's response handling
	// The UserLoginResponse includes AccessTokenCookie and RefreshTokenCookie fields
	// with header:"Set-Cookie" tags that Huma automatically processes
	log.Info().
		Int64("user_id", user.UserID).
		Str("email", input.Body.Email).
		Bool("jwt_development", h.config.JWT.Development).
		Msg("User logged in successfully - creating response with cookies")

	response := models.NewUserLoginResponse(
		accessToken,
		refreshToken, // Include refresh token in response
		int(user.UserID),
		user.Email,
		roles,
		capabilities,
		activePseudonymID,
		displayName,
		pseudonymInfos,
		h.config.JWT.Development,
	)

	log.Info().
		Msg("Created login response with cookies")

	return response, nil
}

// LogoutUser handles user logout
func (h *AuthHandler) LogoutUser(ctx context.Context, input *models.UserLogoutInput) (*models.UserLogoutResponse, error) {
	log.Info().
		Str("endpoint", "auth/logout").
		Str("component", "auth_handler").
		Msg("Processing user logout request")

	// TODO: Implement token blacklisting for logout
	// For now, validate the refresh token if provided (for future blacklisting)
	if input.Body.RefreshToken != "" {
		claims, err := h.validateJWT(input.Body.RefreshToken)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Invalid refresh token provided during logout")
			// Don't return error - still clear cookies even if token is invalid
		} else {
			log.Info().
				Int64("user_id", claims.UserID).
				Str("email", claims.Email).
				Msg("Valid refresh token provided during logout - ready for blacklisting")
			// TODO: Add token to blacklist (Redis/database)
		}
	}

	log.Info().Msg("User logged out successfully - clearing cookies")

	return models.NewUserLogoutResponse(h.config.JWT.Development), nil
}

// validateJWT validates and parses a JWT token
func (h *AuthHandler) validateJWT(tokenString string) (*middleware.JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &middleware.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(*middleware.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, input *models.RefreshTokenInput) (*models.TokenRefreshResponse, error) {
	log.Info().
		Str("endpoint", "auth/refresh").
		Str("component", "auth_handler").
		Msg("Processing token refresh request")

	// Validate the refresh token
	claims, err := h.validateJWT(input.Body.RefreshToken)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Invalid refresh token provided")
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Create user context from the refresh token claims
	userCtx := &middleware.UserContext{
		UserID:            claims.UserID,
		Email:             claims.Email,
		Roles:             claims.Roles,
		Capabilities:      claims.Capabilities,
		MFAEnabled:        claims.MFAEnabled,
		ActivePseudonymID: claims.ActivePseudonymID,
		DisplayName:       claims.DisplayName,
	}

	// Generate new access token
	newAccessToken, err := middleware.GenerateJWT(userCtx, h.config.JWT.Secret, h.config.JWT.Expiration)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", userCtx.UserID).
			Msg("Failed to generate new access token")
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	log.Info().
		Int64("user_id", userCtx.UserID).
		Msg("Token refreshed successfully")

	// Return new token response with cookie
	return models.NewTokenRefreshResponse(newAccessToken, int(h.config.JWT.Expiration.Seconds()), h.config.JWT.Development), nil
}

// hashPassword hashes a password using SHA-256
func (h *AuthHandler) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// verifyPassword verifies a password against a SHA-256 hash
func (h *AuthHandler) verifyPassword(password, hash string) bool {
	// Hash the provided password and compare with stored hash
	passwordHash := h.hashPassword(password)
	return passwordHash == hash
}

// generateSessionToken generates a random session token
func (h *AuthHandler) generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetCurrentUserSession handles getting the current user's session data
func (h *AuthHandler) GetCurrentUserSession(ctx context.Context, input *middleware.AuthInput) (*models.CurrentUserSessionResponse, error) {
	log.Info().
		Str("endpoint", "auth/me").
		Str("component", "auth_handler").
		Msg("Processing get current user session request")

	// Extract user context from the authenticated request
	userCtx, err := middleware.ExtractUserFromHumaInput(input)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "auth/me").Msg("Authentication required for session access")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	userID := int(userCtx.UserID)
	log.Info().
		Int("user_id", userID).
		Str("email", userCtx.Email).
		Msg("Getting current user session data")

	// Get user from database to ensure they still exist and are active
	user, err := h.userDAO.GetUserByID(ctx, int64(userID))
	if err != nil {
		log.Error().Err(err).Int64("user_id", int64(userID)).Msg("Failed to get user from database")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		log.Warn().Int64("user_id", int64(userID)).Msg("User not found")
		return nil, huma.Error404NotFound("User not found")
	}

	// Check if user is active
	if !user.IsActive.Valid || !user.IsActive.V {
		log.Warn().Int64("user_id", int64(userID)).Msg("User account is inactive")
		return nil, huma.Error403Forbidden("Account inactive")
	}

	// Check if user is suspended
	if user.IsSuspended.Valid && user.IsSuspended.V {
		log.Warn().Int64("user_id", int64(userID)).Msg("User account is suspended")
		return nil, huma.Error403Forbidden("Account suspended")
	}

	// Get user roles and capabilities from database
	roles := []string{"user"} // Default role

	// If user has roles stored in database, use those
	if user.Roles.Valid {
		rawValue, err := user.Roles.V.Value()
		if err == nil {
			var userRoles []string
			if err := json.Unmarshal(rawValue.([]byte), &userRoles); err == nil && len(userRoles) > 0 {
				roles = userRoles
			}
		}
	}

	// Get user's pseudonyms for the response
	// Use IBE-based correlation to get user's pseudonyms
	// Use the user's actual roles, not hardcoded "user"
	primaryRole := roles[0] // Use the first role for authentication
	pseudonyms, err := h.securePseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, primaryRole, "authentication")
	if err != nil {
		log.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to get user pseudonyms")
		return nil, fmt.Errorf("failed to get user pseudonyms: %w", err)
	}

	// Convert to API models
	pseudonymInfos := make([]models.PseudonymInfo, len(pseudonyms))
	for i, p := range pseudonyms {
		karmaScore := 0
		if p.KarmaScore.Valid {
			karmaScore = int(p.KarmaScore.V)
		}

		createdAt := time.Now().Format(time.RFC3339)
		if p.CreatedAt.Valid {
			createdAt = p.CreatedAt.V.Format(time.RFC3339)
		}

		lastActiveAt := time.Now().Format(time.RFC3339)
		if p.LastActiveAt.Valid {
			lastActiveAt = p.LastActiveAt.V.Format(time.RFC3339)
		}

		isActive := true
		if p.IsActive.Valid {
			isActive = p.IsActive.V
		}

		pseudonymInfos[i] = models.PseudonymInfo{
			PseudonymID:  p.PseudonymID,
			DisplayName:  p.DisplayName,
			KarmaScore:   karmaScore,
			CreatedAt:    createdAt,
			LastActiveAt: lastActiveAt,
			IsActive:     isActive,
		}
	}

	// Get active pseudonym (use the first one for now, or the one from JWT if available)
	var activePseudonymID string
	var displayName string

	if userCtx.ActivePseudonymID != "" {
		// Use the pseudonym ID from the JWT token
		activePseudonymID = userCtx.ActivePseudonymID
		// Find the display name for this pseudonym
		for _, p := range pseudonymInfos {
			if p.PseudonymID == activePseudonymID {
				displayName = p.DisplayName
				break
			}
		}
	} else if len(pseudonyms) > 0 {
		// Fallback to the first pseudonym
		activePseudonymID = pseudonyms[0].PseudonymID
		displayName = pseudonyms[0].DisplayName
	}

	// Update last active timestamp
	err = h.userDAO.UpdateLastActive(ctx, int64(userID))
	if err != nil {
		log.Error().Err(err).Int64("user_id", int64(userID)).Msg("Failed to update last active timestamp")
		// Don't fail the request for this error
	}

	log.Info().
		Int("user_id", userID).
		Str("email", userCtx.Email).
		Str("active_pseudonym_id", activePseudonymID).
		Msg("Current user session data retrieved successfully")

	return models.NewCurrentUserSessionResponse(
		userID,
		userCtx.Email,
		userCtx.Roles,
		userCtx.Capabilities,
		activePseudonymID,
		displayName,
		pseudonymInfos,
	), nil
}
