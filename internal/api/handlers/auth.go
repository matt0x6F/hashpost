package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/api/validation"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/matt0x6f/hashpost/internal/services"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	config                    *config.Config
	db                        bob.Executor
	userDAO                   dao.UserDAOInterface
	securePseudonymDAO        dao.PseudonymDAOInterface
	identityMappingDAO        dao.IdentityMappingDAOInterface
	roleKeyDAO                dao.RoleKeyDAOInterface
	ibeSystem                 *ibe.IBESystem
	subforumDAO               dao.SubforumDAOInterface
	permissionDAO             dao.PermissionDAOInterface
	emailService              *services.EmailService
	emailVerificationTokenDAO dao.EmailVerificationTokenDAOInterface
	passwordResetTokenDAO     dao.PasswordResetTokenDAOInterface
}

// NewAuthHandler creates a new authentication handler with optional dependencies
// If db is provided, real DAOs will be created. If nil, mock DAOs should be provided.
func NewAuthHandler(
	cfg *config.Config,
	db bob.Executor,
	userDAO dao.UserDAOInterface,
	securePseudonymDAO dao.PseudonymDAOInterface,
	identityMappingDAO dao.IdentityMappingDAOInterface,
	roleKeyDAO dao.RoleKeyDAOInterface,
	ibeSystem *ibe.IBESystem,
	subforumDAO dao.SubforumDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
	emailService *services.EmailService,
	emailVerificationTokenDAO dao.EmailVerificationTokenDAOInterface,
	passwordResetTokenDAO dao.PasswordResetTokenDAOInterface,
) *AuthHandler {
	// If db is provided, create real DAOs (production mode)
	if db != nil {
		userDAO = dao.NewUserDAO(db)
		identityMappingDAO = dao.NewIdentityMappingDAO(db)
		roleKeyDAO = dao.NewRoleKeyDAO(db)
		userBlocksDAO := dao.NewUserBlocksDAO(db)

		// Safe type assertions with error handling
		identityMappingDAOImpl, ok := identityMappingDAO.(*dao.IdentityMappingDAO)
		if !ok {
			log.Error().Msg("identityMappingDAO is not of type *dao.IdentityMappingDAO")
			return nil
		}
		userDAOImpl, ok := userDAO.(*dao.UserDAO)
		if !ok {
			log.Error().Msg("userDAO is not of type *dao.UserDAO")
			return nil
		}
		roleKeyDAOImpl, ok := roleKeyDAO.(*dao.RoleKeyDAO)
		if !ok {
			log.Error().Msg("roleKeyDAO is not of type *dao.RoleKeyDAO")
			return nil
		}

		securePseudonymDAO = dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAOImpl, userDAOImpl, roleKeyDAOImpl, userBlocksDAO)
		subforumDAO = dao.NewSubforumDAO(db)
		permissionDAO = dao.NewPermissionDAO(db)
	}

	return &AuthHandler{
		config:                    cfg,
		db:                        db,
		userDAO:                   userDAO,
		securePseudonymDAO:        securePseudonymDAO,
		identityMappingDAO:        identityMappingDAO,
		roleKeyDAO:                roleKeyDAO,
		ibeSystem:                 ibeSystem,
		subforumDAO:               subforumDAO,
		permissionDAO:             permissionDAO,
		emailService:              emailService,
		emailVerificationTokenDAO: emailVerificationTokenDAO,
		passwordResetTokenDAO:     passwordResetTokenDAO,
	}
}

// generateToken generates a random token for email verification and password reset
func (h *AuthHandler) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(ctx context.Context, input *apimodels.UserRegistrationInput) (*apimodels.UserRegistrationResponse, error) {
	log.Info().
		Str("endpoint", "auth/register").
		Str("component", "auth_handler").
		Msg("Processing user registration request")

	// Enhanced validation using the validation package
	if err := validation.ValidateEmail(input.Body.Email, h.config); err != nil {
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

	// Generate email verification token
	verificationToken, err := h.generateToken()
	if err != nil {
		log.Error().
			Err(err).
			Str("email", input.Body.Email).
			Msg("Failed to generate verification token")
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Create user with email verification pending
	user, err := h.userDAO.CreateUser(ctx, input.Body.Email, hashedPassword)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", input.Body.Email).
			Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Store verification token
	if h.emailVerificationTokenDAO != nil {
		expiresAt := time.Now().Add(24 * time.Hour) // 24 hour expiration
		err = h.emailVerificationTokenDAO.CreateToken(ctx, user.UserID, verificationToken, expiresAt)
		if err != nil {
			log.Error().
				Err(err).
				Int64("user_id", user.UserID).
				Msg("Failed to store verification token")
			// Don't fail registration if token storage fails, just log it
		}
	}

	// Send verification email
	if h.emailService != nil {
		verificationURL := fmt.Sprintf("%s/verify-email?token=%s", h.config.Server.SiteURL, verificationToken)
		err = h.emailService.SendEmail(ctx, "email_verification", input.Body.Email, input.Body.DisplayName, map[string]interface{}{
			"verification_url": verificationURL,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("email", input.Body.Email).
				Msg("Failed to send verification email")
			// Don't fail registration if email fails, just log it
		}
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

	log.Info().
		Int64("user_id", user.UserID).
		Str("email", input.Body.Email).
		Str("pseudonym_id", pseudonym.PseudonymID).
		Msg("User registered successfully")

	// Return registration response without JWT tokens since email verification is required
	return &apimodels.UserRegistrationResponse{
		Status: 200,
		Body: apimodels.UserRegistrationResponseBody{
			UserID:       int(user.UserID),
			Email:        user.Email,
			CreatedAt:    time.Now().Format(time.RFC3339),
			LastActiveAt: time.Now().Format(time.RFC3339),
			IsActive:     true,
			IsSuspended:  false,
			Roles:        roles,
			Capabilities: capabilities,
			PseudonymID:  pseudonym.PseudonymID,
			DisplayName:  pseudonym.DisplayName,
			KarmaScore:   0,
			// No JWT tokens - user must verify email first
			AccessToken:  "",
			RefreshToken: "",
			ExpiresIn:    0,
		},
	}, nil
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(ctx context.Context, input *apimodels.EmailVerificationInput) (*apimodels.EmailVerificationResponse, error) {
	log.Info().
		Str("endpoint", "auth/verify-email").
		Str("component", "auth_handler").
		Msg("Processing email verification request")

	// Get the token from database
	token, err := h.emailVerificationTokenDAO.GetToken(ctx, input.Body.Token)
	if err != nil {
		log.Error().
			Err(err).
			Str("token", input.Body.Token).
			Msg("Failed to get verification token")
		return nil, fmt.Errorf("invalid verification token")
	}

	if token == nil {
		log.Warn().
			Str("token", input.Body.Token).
			Msg("Verification token not found")
		return nil, huma.Error400BadRequest("invalid verification token")
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		log.Warn().
			Str("token", input.Body.Token).
			Time("expires_at", token.ExpiresAt).
			Msg("Verification token expired")
		return nil, huma.Error400BadRequest("verification token expired")
	}

	// Check if token is already used
	if token.UsedAt.Valid {
		log.Warn().
			Str("token", input.Body.Token).
			Msg("Verification token already used")
		return nil, huma.Error400BadRequest("verification token already used")
	}

	// Mark token as used
	err = h.emailVerificationTokenDAO.MarkTokenAsUsed(ctx, input.Body.Token)
	if err != nil {
		log.Error().
			Err(err).
			Str("token", input.Body.Token).
			Msg("Failed to mark token as used")
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Mark user's email as verified
	emailVerified := sql.Null[bool]{}
	emailVerified.Scan(true)

	userUpdates := &models.UserSetter{
		EmailVerified: &emailVerified,
	}

	err = h.userDAO.UpdateUser(ctx, token.UserID, userUpdates)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", token.UserID).
			Msg("Failed to mark user email as verified")
		return nil, fmt.Errorf("failed to mark email as verified: %w", err)
	}

	log.Info().
		Str("token", input.Body.Token).
		Int64("user_id", token.UserID).
		Msg("Email verification completed")

	return &apimodels.EmailVerificationResponse{
		Status: 200,
		Body: apimodels.EmailVerificationResponseBody{
			Message: "Email verified successfully",
		},
	}, nil
}

// RequestPasswordReset handles password reset requests
func (h *AuthHandler) RequestPasswordReset(ctx context.Context, input *apimodels.PasswordResetRequestInput) (*apimodels.PasswordResetRequestResponse, error) {
	log.Info().
		Str("endpoint", "auth/request-password-reset").
		Str("component", "auth_handler").
		Msg("Processing password reset request")

	// Validate email
	if err := validation.ValidateEmail(input.Body.Email, h.config); err != nil {
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}

	// Check if user exists
	user, err := h.userDAO.GetUserByEmail(ctx, input.Body.Email)
	if err != nil || user == nil {
		// Don't reveal if user exists or not for security
		log.Info().
			Str("email", input.Body.Email).
			Msg("Password reset requested for non-existent user")
		return &apimodels.PasswordResetRequestResponse{
			Status: 200,
			Body: apimodels.PasswordResetRequestResponseBody{
				Message: "If an account with this email exists, a password reset link has been sent",
			},
		}, nil
	}

	// Generate reset token
	resetToken, err := h.generateToken()
	if err != nil {
		log.Error().
			Err(err).
			Str("email", input.Body.Email).
			Msg("Failed to generate reset token")
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Store reset token in database with expiration
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour
	err = h.passwordResetTokenDAO.CreateToken(ctx, user.UserID, resetToken, expiresAt)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", input.Body.Email).
			Msg("Failed to store reset token")
		return nil, fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send password reset email
	if h.emailService != nil {
		resetURL := fmt.Sprintf("%s/reset-password/confirm?token=%s", h.config.Server.SiteURL, resetToken)
		err = h.emailService.SendEmail(ctx, "password_reset", input.Body.Email, user.Email, map[string]interface{}{
			"reset_url": resetURL,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("email", input.Body.Email).
				Msg("Failed to send password reset email")
			return nil, fmt.Errorf("failed to send password reset email: %w", err)
		}
	}

	log.Info().
		Str("email", input.Body.Email).
		Msg("Password reset email sent")

	return &apimodels.PasswordResetRequestResponse{
		Status: 200,
		Body: apimodels.PasswordResetRequestResponseBody{
			Message: "If an account with this email exists, a password reset link has been sent",
		},
	}, nil
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(ctx context.Context, input *apimodels.PasswordResetInput) (*apimodels.PasswordResetResponse, error) {
	log.Info().
		Str("endpoint", "auth/reset-password").
		Str("component", "auth_handler").
		Msg("Processing password reset")

	// Validate new password
	if err := validation.ValidatePassword(input.Body.Password, h.config.Security.PasswordValidation); err != nil {
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}

	// Verify and get the reset token
	resetToken, err := h.passwordResetTokenDAO.GetToken(ctx, input.Body.Token)
	if err != nil {
		log.Error().
			Err(err).
			Str("token", input.Body.Token).
			Msg("Failed to retrieve reset token")
		return nil, fmt.Errorf("failed to verify reset token: %w", err)
	}

	if resetToken == nil {
		return nil, huma.Error400BadRequest("Invalid or expired reset token")
	}

	// Check if token has expired
	if time.Now().After(resetToken.ExpiresAt) {
		return nil, huma.Error400BadRequest("Reset token has expired")
	}

	// Check if token has already been used
	if resetToken.UsedAt.Valid {
		return nil, huma.Error400BadRequest("Reset token has already been used")
	}

	// Hash the new password
	hashedPassword := h.hashPassword(input.Body.Password)

	// Update user's password
	userUpdates := &models.UserSetter{
		PasswordHash: &hashedPassword,
	}
	err = h.userDAO.UpdateUser(ctx, resetToken.UserID, userUpdates)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", resetToken.UserID).
			Msg("Failed to update user password")
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	err = h.passwordResetTokenDAO.MarkTokenAsUsed(ctx, input.Body.Token)
	if err != nil {
		log.Error().
			Err(err).
			Str("token", input.Body.Token).
			Msg("Failed to mark reset token as used")
		// Don't fail the request if marking as used fails
	}

	log.Info().
		Str("token", input.Body.Token).
		Int64("user_id", resetToken.UserID).
		Msg("Password reset completed")

	return &apimodels.PasswordResetResponse{
		Status: 200,
		Body: apimodels.PasswordResetResponseBody{
			Message: "Password reset successfully",
		},
	}, nil
}

// LoginUser handles user login
func (h *AuthHandler) LoginUser(ctx context.Context, input *apimodels.UserLoginInput) (*apimodels.UserLoginResponse, error) {
	log.Info().
		Str("endpoint", "auth/login").
		Str("component", "auth_handler").
		Msg("Processing user login request")

	// Enhanced validation using the validation package
	if err := validation.ValidateEmail(input.Body.Email, h.config); err != nil {
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

	// Check if email is verified
	if !user.EmailVerified.Valid || !user.EmailVerified.V {
		log.Warn().
			Int64("user_id", user.UserID).
			Str("email", user.Email).
			Msg("User attempted to login with unverified email")
		return nil, fmt.Errorf("email not verified. Please check your email and click the verification link before logging in")
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
	pseudonymInfos := make([]apimodels.PseudonymInfo, len(pseudonyms))
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

		pseudonymInfos[i] = apimodels.PseudonymInfo{
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

	response := apimodels.NewUserLoginResponse(
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
func (h *AuthHandler) LogoutUser(ctx context.Context, input *apimodels.UserLogoutInput) (*apimodels.UserLogoutResponse, error) {
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

	return apimodels.NewUserLogoutResponse(h.config.JWT.Development), nil
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
	Body         apimodels.RefreshTokenBody
}) (*apimodels.TokenRefreshResponse, error) {
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
		return nil, huma.Error401Unauthorized("Invalid or missing refresh token")
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
	return apimodels.NewTokenRefreshResponse(newAccessToken, int(h.config.JWT.Expiration.Seconds()), h.config.JWT.Development), nil
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
func (h *AuthHandler) GetCurrentUserSession(ctx context.Context, input *middleware.AuthInput) (*apimodels.CurrentUserSessionResponse, error) {
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
	pseudonymInfos := make([]apimodels.PseudonymInfo, len(pseudonyms))
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

		pseudonymInfos[i] = apimodels.PseudonymInfo{
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

	// Get active pseudonym's roles and capabilities (global only, no subforum context)
	activeRoles, activeCapabilities, err := h.permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(ctx, int64(userID), activePseudonymID, nil)
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

	return apimodels.NewCurrentUserSessionResponse(
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
}) (*apimodels.CurrentUserSessionResponse, error) {
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

	// Parse subforum name to extract community type and actual name
	communityType, subforumName, err := h.parseSubforumName(input.SubforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", input.SubforumName).Msg("Failed to parse subforum name")
		return nil, fmt.Errorf("invalid subforum name format: %w", err)
	}

	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, communityType, subforumName)
	if err == nil && subforum != nil {
		// Check if the active pseudonym has moderator capabilities for this subforum
		activePseudonymID := userCtx.ActivePseudonymID
		if activePseudonymID != "" {
			log.Debug().
				Int64("user_id", int64(userID)).
				Int32("subforum_id", subforum.SubforumID).
				Str("active_pseudonym_id", activePseudonymID).
				Msg("Checking moderator capabilities for active pseudonym")

			// Use the unified permission system to get roles and capabilities
			subforumID := &subforum.SubforumID
			unifiedRoles, unifiedCapabilities, err := h.permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(ctx, int64(userID), activePseudonymID, subforumID)
			if err != nil {
				log.Error().Err(err).Int64("user_id", int64(userID)).Str("active_pseudonym_id", activePseudonymID).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get unified roles and capabilities")
				return nil, fmt.Errorf("failed to get unified roles and capabilities: %w", err)
			}

			// Check if the active pseudonym is the owner of this subforum
			isOwner := subforum.OwnerPseudonymID.Valid && subforum.OwnerPseudonymID.V == activePseudonymID
			if isOwner {
				log.Debug().
					Int64("user_id", int64(userID)).
					Int32("subforum_id", subforum.SubforumID).
					Str("active_pseudonym_id", activePseudonymID).
					Msg("Found owner record - granting owner and moderator capabilities")

				// Add owner role
				if !contains(unifiedRoles, "owner") {
					unifiedRoles = append(unifiedRoles, "owner")
				}

				// Add owner capabilities (all moderator capabilities plus owner-specific ones)
				unifiedCapabilities = append(unifiedCapabilities, "moderate_content", "ban_users", "manage_moderators", "sticky_post", "lock_post")
			}

			// Use the unified roles and capabilities
			roles = unifiedRoles
			subforumCapabilities = unifiedCapabilities
		}
	}

	// Use the unified roles and capabilities (they already include both global and subforum-specific)
	allRoles := roles
	allCapabilities := subforumCapabilities

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
	pseudonymInfos := make([]apimodels.PseudonymInfo, len(pseudonyms))
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

		pseudonymInfos[i] = apimodels.PseudonymInfo{
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

	return apimodels.NewCurrentUserSessionResponse(
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
	apimodels.SwitchPseudonymInput
}) (*apimodels.SwitchPseudonymResponse, error) {
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

	return apimodels.NewSwitchPseudonymResponse(accessToken, h.config.JWT.Expiration, h.config.JWT.Development), nil
}

// DeactivatePseudonym handles deactivating a pseudonym owned by the current user
func (h *AuthHandler) DeactivatePseudonym(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.DeactivatePseudonymInput
}) (*apimodels.DeactivatePseudonymResponse, error) {
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

	return apimodels.NewDeactivatePseudonymResponse(), nil
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

// parseSubforumName parses a full subforum name (e.g., "t/subforum-name") into community type and name
func (h *AuthHandler) parseSubforumName(fullName string) (communityType, subforumName string, err error) {
	// Handle different formats:
	// 1. "t/subforum-name" -> communityType: "t", subforumName: "subforum-name"
	// 2. "subforum-name" -> communityType: "h", subforumName: "subforum-name" (default for h/ subforums)

	if fullName == "" {
		return "", "", fmt.Errorf("subforum name cannot be empty")
	}

	// Check if it contains a slash (community type prefix)
	if strings.Contains(fullName, "/") {
		parts := strings.SplitN(fullName, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid subforum name format: expected 'community-type/name'")
		}

		communityType = parts[0]
		subforumName = parts[1]

		// Validate community type
		validTypes := []string{"t", "g", "b", "c", "h"}
		isValid := false
		for _, validType := range validTypes {
			if communityType == validType {
				isValid = true
				break
			}
		}

		if !isValid {
			return "", "", fmt.Errorf("invalid community type: %s", communityType)
		}

		return communityType, subforumName, nil
	}

	// No slash found, treat as h/ subforum (default)
	return "h", fullName, nil
}
