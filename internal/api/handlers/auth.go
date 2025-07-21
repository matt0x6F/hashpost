package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/api/constants"
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
	userDAO            dao.UserDAOInterface
	securePseudonymDAO dao.SecurePseudonymDAOInterface
	identityMappingDAO dao.IdentityMappingDAOInterface
	roleKeyDAO         dao.RoleKeyDAOInterface
	ibeSystem          *ibe.IBESystem
	subforumDAO        dao.SubforumDAOInterface
	permissionDAO      dao.PermissionDAOInterface
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(cfg *config.Config, db bob.Executor, rawDB *sql.DB) *AuthHandler {
	userDAO := dao.NewUserDAO(db)
	ibeSystem := ibe.NewIBESystem()
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	securePseudonymDAO := dao.NewSecurePseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)
	subforumDAO := dao.NewSubforumDAO(db)
	permissionDAO := dao.NewPermissionDAO(db)

	return &AuthHandler{
		config:             cfg,
		db:                 db,
		userDAO:            userDAO,
		securePseudonymDAO: securePseudonymDAO,
		identityMappingDAO: identityMappingDAO,
		ibeSystem:          ibeSystem,
		subforumDAO:        subforumDAO,
		permissionDAO:      permissionDAO,
	}
}

// NewAuthHandlerWithIBE creates a new authentication handler with a specific IBE system
func NewAuthHandlerWithIBE(cfg *config.Config, db bob.Executor, rawDB *sql.DB, ibeSystem *ibe.IBESystem) *AuthHandler {
	userDAO := dao.NewUserDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	securePseudonymDAO := dao.NewSecurePseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)
	subforumDAO := dao.NewSubforumDAO(db)
	permissionDAO := dao.NewPermissionDAO(db)

	return &AuthHandler{
		config:             cfg,
		db:                 db,
		userDAO:            userDAO,
		securePseudonymDAO: securePseudonymDAO,
		identityMappingDAO: identityMappingDAO,
		ibeSystem:          ibeSystem,
		subforumDAO:        subforumDAO,
		permissionDAO:      permissionDAO,
	}
}

// NewAuthHandlerWithDependencies creates a new authentication handler with injected dependencies
func NewAuthHandlerWithDependencies(cfg *config.Config, userDAO dao.UserDAOInterface, securePseudonymDAO dao.SecurePseudonymDAOInterface, identityMappingDAO dao.IdentityMappingDAOInterface, roleKeyDAO dao.RoleKeyDAOInterface, ibeSystem *ibe.IBESystem, subforumDAO dao.SubforumDAOInterface, permissionDAO dao.PermissionDAOInterface) *AuthHandler {
	return &AuthHandler{
		config:             cfg,
		db:                 nil, // Will be set by individual constructors
		userDAO:            userDAO,
		securePseudonymDAO: securePseudonymDAO,
		identityMappingDAO: identityMappingDAO,
		roleKeyDAO:         roleKeyDAO,
		ibeSystem:          ibeSystem,
		subforumDAO:        subforumDAO,
		permissionDAO:      permissionDAO,
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
	if err := h.roleKeyDAO.EnsureDefaultKeys(ctx, h.ibeSystem, user.UserID); err != nil {
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

	// Get pseudonym capabilities
	if pseudonym.Capabilities.Valid {
		rawValue, err := pseudonym.Capabilities.V.Value()
		if err == nil {
			var pseudonymCapabilities []string
			if err := json.Unmarshal(rawValue.([]byte), &pseudonymCapabilities); err == nil && len(pseudonymCapabilities) > 0 {
				capabilities = pseudonymCapabilities
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

	// Note: Role keys are created during user registration, not during login

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

	// Note: Capabilities will be set from the default pseudonym later in the method

	// Get user's pseudonyms for the response
	// Use IBE-based correlation to get user's pseudonyms
	// Use the user's actual roles, not hardcoded "user"
	primaryRole := roles[0] // Use the first role for authentication
	pseudonyms, err := h.securePseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, primaryRole, constants.ScopeAuthentication)
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
	defaultPseudonym, err := h.securePseudonymDAO.GetDefaultPseudonymByUserID(ctx, user.UserID, primaryRole, constants.ScopeAuthentication)
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
func (h *AuthHandler) RefreshToken(ctx context.Context, input *struct {
	RefreshToken string `cookie:"refresh_token"`
	Body         models.RefreshTokenBody
}) (*models.TokenRefreshResponse, error) {
	log.Info().
		Str("endpoint", "auth/refresh").
		Str("component", "auth_handler").
		Msg("Processing token refresh request")

	// Prefer the cookie, fall back to body for non-browser clients
	refreshToken := input.RefreshToken
	if refreshToken == "" {
		refreshToken = input.Body.RefreshToken
	}

	// Validate the refresh token
	claims, err := h.validateJWT(refreshToken)
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

	// Get user roles from database
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

	// Note: Capabilities will be set from the active pseudonym later in the method

	// Get user's pseudonyms for the response
	// Use IBE-based correlation to get user's pseudonyms
	// Use the user's actual roles, not hardcoded "user"
	primaryRole := roles[0] // Use the first role for authentication
	pseudonyms, err := h.securePseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, primaryRole, constants.ScopeAuthentication)
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

	// Get active pseudonym's roles and capabilities
	activeRoles, activeCapabilities, err := h.permissionDAO.GetActivePseudonymRolesAndCapabilities(ctx, int64(userID), activePseudonymID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", int64(userID)).Str("active_pseudonym_id", activePseudonymID).Msg("Failed to get active pseudonym roles and capabilities")
		return nil, fmt.Errorf("failed to get active pseudonym roles and capabilities: %w", err)
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
		activeRoles,
		activeCapabilities,
		activePseudonymID,
		displayName,
		pseudonymInfos,
	), nil
}

// GetCurrentUserSessionForSubforum handles getting the current user's session data with subforum-specific capabilities
func (h *AuthHandler) GetCurrentUserSessionForSubforum(ctx context.Context, input *struct {
	middleware.AuthInput
	SubforumName string `path:"subforum_name"`
}) (*models.CurrentUserSessionResponse, error) {
	log.Info().
		Str("endpoint", "auth/me/subforum").
		Str("component", "auth_handler").
		Str("subforum_name", input.SubforumName).
		Msg("Processing get current user session for subforum request")

	// Extract user context from the authenticated request
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "auth/me/subforum").Msg("Authentication required for session access")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	userID := int(userCtx.UserID)
	log.Info().
		Int("user_id", userID).
		Str("email", userCtx.Email).
		Str("subforum_name", input.SubforumName).
		Msg("Getting current user session data for subforum")

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

	// Get subforum-specific capabilities
	subforumCapabilities := []string{}

	// Get subforum by name
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, input.SubforumName)
	if err == nil && subforum != nil {
		// Check if the active pseudonym has moderator capabilities for this subforum
		activePseudonymID := userCtx.ActivePseudonymID
		if activePseudonymID != "" {
			log.Debug().
				Int64("user_id", int64(userID)).
				Int32("subforum_id", subforum.SubforumID).
				Str("active_pseudonym_id", activePseudonymID).
				Msg("Checking moderator capabilities for active pseudonym")

			// Check if the active pseudonym is a moderator for this subforum
			// Use the permission DAO to check moderator status
			hasModerateContent, err := h.permissionDAO.HasSubforumCapabilityWithActivePseudonym(ctx, int64(userID), subforum.SubforumID, "moderate_content", activePseudonymID)
			if err == nil && hasModerateContent {
				log.Debug().
					Int64("user_id", int64(userID)).
					Int32("subforum_id", subforum.SubforumID).
					Str("active_pseudonym_id", activePseudonymID).
					Msg("Found moderator record")

				// Only add moderator role if the user is not a platform admin
				// Platform admins have moderator capabilities but shouldn't get the moderator role
				if !contains(userCtx.Roles, "platform_admin") {
					// Add moderator role to roles array
					if !contains(roles, "moderator") {
						roles = append(roles, "moderator")
					}
				}

				// Add moderator capabilities
				subforumCapabilities = append(subforumCapabilities, "moderate_content")

				// Check for additional moderator capabilities
				hasBanUsers, _ := h.permissionDAO.HasSubforumCapabilityWithActivePseudonym(ctx, int64(userID), subforum.SubforumID, "ban_users", activePseudonymID)
				if hasBanUsers {
					subforumCapabilities = append(subforumCapabilities, "ban_users")
				}

				hasManageModerators, _ := h.permissionDAO.HasSubforumCapabilityWithActivePseudonym(ctx, int64(userID), subforum.SubforumID, "manage_moderators", activePseudonymID)
				if hasManageModerators {
					subforumCapabilities = append(subforumCapabilities, "manage_moderators")
				}
			} else {
				log.Debug().
					Int64("user_id", int64(userID)).
					Int32("subforum_id", subforum.SubforumID).
					Str("active_pseudonym_id", activePseudonymID).
					Msg("No moderator record found")
			}
		}
	}

	// Use roles and capabilities from the user context (JWT token)
	pseudonymRoles := userCtx.Roles
	pseudonymCapabilities := userCtx.Capabilities

	// Combine pseudonym roles with subforum-specific roles
	allRoles := append(pseudonymRoles, roles...)
	allRoles = removeDuplicates(allRoles)

	// Combine pseudonym capabilities with subforum-specific capabilities
	allCapabilities := append(pseudonymCapabilities, subforumCapabilities...)

	// Get user's pseudonyms for the response
	// Use IBE-based correlation to get user's pseudonyms
	// Use the user's actual roles, not hardcoded "user"
	primaryRole := roles[0] // Use the first role for authentication
	pseudonyms, err := h.securePseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, primaryRole, constants.ScopeAuthentication)
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
		Str("subforum_name", input.SubforumName).
		Int("subforum_capabilities", len(subforumCapabilities)).
		Msg("Current user session data for subforum retrieved successfully")

	return models.NewCurrentUserSessionResponse(
		userID,
		userCtx.Email,
		allRoles,        // Use the updated roles array that includes subforum-specific roles
		allCapabilities, // Include subforum-specific capabilities
		activePseudonymID,
		displayName,
		pseudonymInfos,
	), nil
}

// SwitchPseudonym handles switching the user's active pseudonym
func (h *AuthHandler) SwitchPseudonym(ctx context.Context, input *struct {
	middleware.AuthInput
	models.SwitchPseudonymInput
}) (*models.SwitchPseudonymResponse, error) {
	log.Info().
		Str("endpoint", "auth/switch-pseudonym").
		Str("component", "auth_handler").
		Msg("Processing pseudonym switch request")

	// Extract user context from JWT token
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	// Validate input
	if input.Body.PseudonymID == "" {
		return nil, huma.Error400BadRequest("Pseudonym ID is required")
	}

	// Check if user is trying to switch to the same pseudonym
	if userCtx.ActivePseudonymID == input.Body.PseudonymID {
		return nil, huma.Error400BadRequest("Already using this pseudonym")
	}

	// Get the target pseudonym
	targetPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, input.Body.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", input.Body.PseudonymID).Msg("Failed to get target pseudonym")
		return nil, huma.Error404NotFound("Target pseudonym not found")
	}
	if targetPseudonym == nil {
		return nil, huma.Error404NotFound("Target pseudonym not found")
	}

	// Verify ownership using multi-scope fallback strategy
	// Try authentication scope first (most secure), then self-correlation scope
	ownsPseudonym := false
	var ownershipErr error

	// Try each role with authentication scope first
	for _, role := range userCtx.Roles {
		ownsPseudonym, ownershipErr = h.securePseudonymDAO.VerifyPseudonymOwnership(ctx, input.Body.PseudonymID, userCtx.UserID, role, constants.ScopeAuthentication)
		if ownershipErr == nil && ownsPseudonym {
			break
		}
	}

	// If authentication scope fails, try self-correlation scope
	if !ownsPseudonym {
		for _, role := range userCtx.Roles {
			ownsPseudonym, ownershipErr = h.securePseudonymDAO.VerifyPseudonymOwnership(ctx, input.Body.PseudonymID, userCtx.UserID, role, constants.ScopeSelfCorrelation)
			if ownershipErr == nil && ownsPseudonym {
				break
			}
		}
	}

	if !ownsPseudonym {
		log.Warn().
			Int64("user_id", userCtx.UserID).
			Str("target_pseudonym_id", input.Body.PseudonymID).
			Msg("User attempted to switch to pseudonym they don't own")
		return nil, huma.Error403Forbidden("You do not own this pseudonym")
	}

	// Update last active timestamp for the target pseudonym
	err = h.securePseudonymDAO.UpdateLastActive(ctx, input.Body.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", input.Body.PseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	// Generate new JWT token with updated pseudonym context
	newUserCtx := &middleware.UserContext{
		UserID:            userCtx.UserID,
		Email:             userCtx.Email,
		Roles:             userCtx.Roles,
		Capabilities:      userCtx.Capabilities,
		ActivePseudonymID: input.Body.PseudonymID,
		DisplayName:       targetPseudonym.DisplayName,
	}

	accessToken, err := middleware.GenerateJWT(newUserCtx, h.config.JWT.Secret, h.config.JWT.Expiration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate new JWT token")
		return nil, huma.Error500InternalServerError("Failed to generate new token")
	}

	log.Info().
		Int64("user_id", userCtx.UserID).
		Str("old_pseudonym_id", userCtx.ActivePseudonymID).
		Str("new_pseudonym_id", input.Body.PseudonymID).
		Msg("Pseudonym switched successfully")

	return models.NewSwitchPseudonymResponse(accessToken, h.config.JWT.Expiration, h.config.JWT.Development), nil
}

// DeactivatePseudonym handles deactivating a pseudonym owned by the current user
func (h *AuthHandler) DeactivatePseudonym(ctx context.Context, input *struct {
	middleware.AuthInput
	models.DeactivatePseudonymInput
}) (*models.DeactivatePseudonymResponse, error) {
	log.Info().
		Str("endpoint", "auth/deactivate-pseudonym").
		Str("component", "auth_handler").
		Msg("Processing pseudonym deactivation request")

	// Extract user context from JWT token
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	// Validate input
	if input.Body.PseudonymID == "" {
		return nil, huma.Error400BadRequest("Pseudonym ID is required")
	}

	// Check if user is trying to deactivate their active pseudonym
	if userCtx.ActivePseudonymID == input.Body.PseudonymID {
		return nil, huma.Error400BadRequest("Cannot deactivate your active pseudonym. Please switch to a different pseudonym first.")
	}

	// Deactivate the pseudonym using self-correlation scope (user can only deactivate their own pseudonyms)
	err = h.securePseudonymDAO.DeactivatePseudonym(ctx, input.Body.PseudonymID, userCtx.UserID, "user", constants.ScopeSelfCorrelation)
	if err != nil {
		log.Error().Err(err).
			Int64("user_id", userCtx.UserID).
			Str("pseudonym_id", input.Body.PseudonymID).
			Msg("Failed to deactivate pseudonym")

		// Return appropriate error based on the failure reason
		if err.Error() == "not found" {
			return nil, huma.Error404NotFound("Pseudonym not found")
		}
		if err.Error() == "does not own" {
			return nil, huma.Error403Forbidden("You do not own this pseudonym")
		}
		return nil, huma.Error500InternalServerError("Failed to deactivate pseudonym")
	}

	log.Info().
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", input.Body.PseudonymID).
		Msg("Pseudonym deactivated successfully")

	return models.NewDeactivatePseudonymResponse(), nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			result = append(result, item)
			seen[item] = struct{}{}
		}
	}
	return result
}
