package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stephenafamo/bob"
)

// MockPermissionDAO is a mock implementation for testing
type MockPermissionDAO struct {
	hasSubforumCapabilityFunc       func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error)
	canAccessPrivateSubforumFunc    func(ctx context.Context, userID int64, subforumID int32) (bool, error)
	getUserSubforumRolesFunc        func(ctx context.Context, userID int64, subforumID int32) ([]string, error)
	getUserSubforumCapabilitiesFunc func(ctx context.Context, userID int64, subforumID int32) ([]string, error)
}

func (m *MockPermissionDAO) HasSubforumCapability(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
	if m.hasSubforumCapabilityFunc != nil {
		return m.hasSubforumCapabilityFunc(ctx, userID, subforumID, capability)
	}
	return false, nil
}

func (m *MockPermissionDAO) CanAccessPrivateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	if m.canAccessPrivateSubforumFunc != nil {
		return m.canAccessPrivateSubforumFunc(ctx, userID, subforumID)
	}
	return false, nil
}

func (m *MockPermissionDAO) GetUserSubforumRoles(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	if m.getUserSubforumRolesFunc != nil {
		return m.getUserSubforumRolesFunc(ctx, userID, subforumID)
	}
	return []string{}, nil
}

func (m *MockPermissionDAO) GetUserSubforumCapabilities(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	if m.getUserSubforumCapabilitiesFunc != nil {
		return m.getUserSubforumCapabilitiesFunc(ctx, userID, subforumID)
	}
	return []string{}, nil
}

// MockPermissionMiddleware wraps PermissionMiddleware with mock DAO
type MockPermissionMiddleware struct {
	*PermissionMiddleware
	mockDAO *MockPermissionDAO
}

func NewMockPermissionMiddleware() *MockPermissionMiddleware {
	mockDAO := &MockPermissionDAO{}
	return &MockPermissionMiddleware{
		PermissionMiddleware: NewPermissionMiddlewareWithDAO(mockDAO),
		mockDAO:              mockDAO,
	}
}

func TestRequireSubforumCapability_Success(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock successful capability check
	mockMiddleware.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		return true, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Check that subforum_id is in context
		subforumID, ok := GetSubforumIDFromContext(r)
		if !ok {
			t.Error("Expected subforum_id in context")
		}
		if subforumID != 123 {
			t.Errorf("Expected subforum_id 123, got %d", subforumID)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := mockMiddleware.RequireSubforumCapability("moderate_content")

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireSubforumCapability_Unauthorized(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create middleware
	middleware := mockMiddleware.RequireSubforumCapability("moderate_content")

	// Create request without user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRequireSubforumCapability_MissingSubforumID(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create middleware
	middleware := mockMiddleware.RequireSubforumCapability("moderate_content")

	// Create request without subforum_id
	req := httptest.NewRequest("GET", "/test", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRequireSubforumCapability_InvalidSubforumID(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create middleware
	middleware := mockMiddleware.RequireSubforumCapability("moderate_content")

	// Create request with invalid subforum_id
	req := httptest.NewRequest("GET", "/test?subforum_id=invalid", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRequireSubforumCapability_Forbidden(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock failed capability check
	mockMiddleware.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		return false, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create middleware
	middleware := mockMiddleware.RequireSubforumCapability("moderate_content")

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequireSubforumCapability_DatabaseError(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock database error
	mockMiddleware.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		return false, context.DeadlineExceeded
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create middleware
	middleware := mockMiddleware.RequireSubforumCapability("moderate_content")

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRequirePrivateSubforumAccess_Success(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock successful access check
	mockMiddleware.mockDAO.canAccessPrivateSubforumFunc = func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
		return true, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Check that subforum_id is in context
		subforumID, ok := GetSubforumIDFromContext(r)
		if !ok {
			t.Error("Expected subforum_id in context")
		}
		if subforumID != 123 {
			t.Errorf("Expected subforum_id 123, got %d", subforumID)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := mockMiddleware.RequirePrivateSubforumAccess()

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequirePrivateSubforumAccess_Forbidden(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock failed access check
	mockMiddleware.mockDAO.canAccessPrivateSubforumFunc = func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
		return false, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create middleware
	middleware := mockMiddleware.RequirePrivateSubforumAccess()

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequireModerationCapability(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock successful capability check
	mockMiddleware.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		if capability == "moderate_content" {
			return true, nil
		}
		return false, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := mockMiddleware.RequireModerationCapability()

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireBanCapability(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock successful capability check
	mockMiddleware.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		if capability == "ban_users" {
			return true, nil
		}
		return false, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := mockMiddleware.RequireBanCapability()

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireRemoveContentCapability(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock successful capability check
	mockMiddleware.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		if capability == "remove_content" {
			return true, nil
		}
		return false, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := mockMiddleware.RequireRemoveContentCapability()

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireManageModeratorsCapability(t *testing.T) {
	mockMiddleware := NewMockPermissionMiddleware()

	// Mock successful capability check
	mockMiddleware.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		if capability == "manage_moderators" {
			return true, nil
		}
		return false, nil
	}

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := mockMiddleware.RequireManageModeratorsCapability()

	// Create request with user context
	req := httptest.NewRequest("GET", "/test?subforum_id=123", nil)
	userCtx := &UserContext{
		UserID: 456,
		Email:  "test@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute middleware
	middleware(testHandler).ServeHTTP(w, req)

	// Verify results
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetSubforumIDFromContext(t *testing.T) {
	// Test with subforum_id in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "subforum_id", int32(123))
	req = req.WithContext(ctx)

	subforumID, ok := GetSubforumIDFromContext(req)
	if !ok {
		t.Error("Expected to find subforum_id in context")
	}
	if subforumID != 123 {
		t.Errorf("Expected subforum_id 123, got %d", subforumID)
	}

	// Test without subforum_id in context
	req2 := httptest.NewRequest("GET", "/test", nil)
	subforumID2, ok2 := GetSubforumIDFromContext(req2)
	if ok2 {
		t.Error("Expected not to find subforum_id in context")
	}
	if subforumID2 != 0 {
		t.Errorf("Expected subforum_id 0, got %d", subforumID2)
	}
}

// MockPermissionChecker wraps PermissionChecker with mock DAO
type MockPermissionChecker struct {
	*PermissionChecker
	mockDAO *MockPermissionDAO
}

func NewMockPermissionChecker() *MockPermissionChecker {
	mockDAO := &MockPermissionDAO{}
	return &MockPermissionChecker{
		PermissionChecker: NewPermissionCheckerWithDAO(mockDAO),
		mockDAO:           mockDAO,
	}
}

func TestPermissionChecker_CheckSubforumCapability(t *testing.T) {
	mockChecker := NewMockPermissionChecker()

	// Mock successful capability check
	mockChecker.mockDAO.hasSubforumCapabilityFunc = func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
		return true, nil
	}

	ctx := context.Background()
	hasCapability, err := mockChecker.CheckSubforumCapability(ctx, 123, 456, "moderate_content")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !hasCapability {
		t.Error("Expected capability check to return true")
	}
}

func TestPermissionChecker_CheckPrivateSubforumAccess(t *testing.T) {
	mockChecker := NewMockPermissionChecker()

	// Mock successful access check
	mockChecker.mockDAO.canAccessPrivateSubforumFunc = func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
		return true, nil
	}

	ctx := context.Background()
	canAccess, err := mockChecker.CheckPrivateSubforumAccess(ctx, 123, 456)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !canAccess {
		t.Error("Expected access check to return true")
	}
}

func TestPermissionChecker_GetUserSubforumRoles(t *testing.T) {
	mockChecker := NewMockPermissionChecker()

	expectedRoles := []string{"moderator", "admin"}
	mockChecker.mockDAO.getUserSubforumRolesFunc = func(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
		return expectedRoles, nil
	}

	ctx := context.Background()
	roles, err := mockChecker.GetUserSubforumRoles(ctx, 123, 456)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(roles) != len(expectedRoles) {
		t.Errorf("Expected %d roles, got %d", len(expectedRoles), len(roles))
	}
	for i, role := range roles {
		if role != expectedRoles[i] {
			t.Errorf("Expected role %s, got %s", expectedRoles[i], role)
		}
	}
}

func TestPermissionChecker_GetUserSubforumCapabilities(t *testing.T) {
	mockChecker := NewMockPermissionChecker()

	expectedCapabilities := []string{"moderate_content", "ban_users", "remove_content"}
	mockChecker.mockDAO.getUserSubforumCapabilitiesFunc = func(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
		return expectedCapabilities, nil
	}

	ctx := context.Background()
	capabilities, err := mockChecker.GetUserSubforumCapabilities(ctx, 123, 456)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(capabilities) != len(expectedCapabilities) {
		t.Errorf("Expected %d capabilities, got %d", len(expectedCapabilities), len(capabilities))
	}
	for i, capability := range capabilities {
		if capability != expectedCapabilities[i] {
			t.Errorf("Expected capability %s, got %s", expectedCapabilities[i], capability)
		}
	}
}

func TestNewPermissionMiddleware(t *testing.T) {
	// This test verifies that the middleware can be created with a mock database
	var mockDB bob.Executor = nil // In real tests, you'd use a proper mock

	middleware := NewPermissionMiddleware(mockDB)
	if middleware == nil {
		t.Error("Expected middleware to be created")
	}
	if middleware.permissionDAO == nil {
		t.Error("Expected permission DAO to be initialized")
	}
}

func TestNewPermissionChecker(t *testing.T) {
	// This test verifies that the permission checker can be created with a mock database
	var mockDB bob.Executor = nil // In real tests, you'd use a proper mock

	checker := NewPermissionChecker(mockDB)
	if checker == nil {
		t.Error("Expected permission checker to be created")
	}
	if checker.permissionDAO == nil {
		t.Error("Expected permission DAO to be initialized")
	}
}
