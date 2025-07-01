//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/testutil"
)

func TestAuthLogin_Integration(t *testing.T) {
	t.Run("LoginWithValidCredentials", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, testutil.GenerateUniqueEmail("login_valid"), "TestPassword123!", []string{"user"})
		resp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		token := suite.ExtractTokenFromResponse(t, resp)
		if token == "" {
			t.Error("Expected non-empty access token")
		}
		var response models.UserLoginResponseBody
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if response.Email != testUser.Email {
			t.Errorf("Expected email %s, got %s", testUser.Email, response.Email)
		}
		if response.AccessToken == "" {
			t.Error("Expected non-empty access_token in response")
		}
		if response.RefreshToken == "" {
			t.Error("Expected non-empty refresh_token in response")
		}
	})

	t.Run("LoginWithInvalidCredentials", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		resp := suite.LoginUser(t, server, "nonexistent@example.com", "WrongPassword")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for invalid credentials, got %d", resp.StatusCode)
		}
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if status, ok := responseBody["status"].(float64); !ok || int(status) != 500 {
			t.Error("Expected status 500 in error response")
		}
		if detail, ok := responseBody["detail"].(string); !ok || detail == "" {
			t.Error("Expected non-empty detail in error response")
		}
	})

	t.Run("LoginWithWrongPassword", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, testutil.GenerateUniqueEmail("login_wrong_pass"), "CorrectPassword123!", []string{"user"})
		resp := suite.LoginUser(t, server, testUser.Email, "WrongPassword")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for wrong password, got %d", resp.StatusCode)
		}
	})

	t.Run("LoginWithEmptyEmail", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		resp := suite.LoginUser(t, server, "", "SomePassword")
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Expected status 422 for empty email, got %d", resp.StatusCode)
		}
	})

	t.Run("LoginWithEmptyPassword", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, testutil.GenerateUniqueEmail("login_empty_pass"), "CorrectPassword123!", []string{"user"})
		resp := suite.LoginUser(t, server, testUser.Email, "")
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Expected status 422 for empty password, got %d", resp.StatusCode)
		}
	})
}

func TestAuthLogin_AdminUser_Integration(t *testing.T) {
	t.Run("LoginWithAdminUser", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "admin@example.com", "AdminPassword123!", []string{"platform_admin"})
		resp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		token := suite.ExtractTokenFromResponse(t, resp)
		if token == "" {
			t.Error("Expected non-empty access token")
		}
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if email, ok := responseBody["email"].(string); !ok || email != testUser.Email {
			t.Errorf("Expected email %s, got %s", testUser.Email, email)
		}
		if accessToken, ok := responseBody["access_token"].(string); !ok || accessToken == "" {
			t.Error("Expected non-empty access_token in response")
		}
		if refreshToken, ok := responseBody["refresh_token"].(string); !ok || refreshToken == "" {
			t.Error("Expected non-empty refresh_token in response")
		}
	})
}

func TestAuthLogin_JSONParsing_Integration(t *testing.T) {
	t.Run("LoginWithValidJSON", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "json@example.com", "JsonPassword123!", []string{"user"})
		loginData := models.UserLoginBody{
			Email:    testUser.Email,
			Password: testUser.Password,
		}
		jsonData, err := json.Marshal(loginData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if email, ok := responseBody["email"].(string); !ok || email != testUser.Email {
			t.Errorf("Expected email %s, got %s", testUser.Email, email)
		}
	})

	t.Run("LoginWithInvalidJSON", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBufferString(`{"email": "test@example.com", "password": "test"`))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		if resp.StatusCode < 400 {
			t.Errorf("Expected error status for invalid JSON, got %d", resp.StatusCode)
		}
	})
}

func TestAuthLogin_Debug_Integration(t *testing.T) {
	t.Run("DebugLoginRequest", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "debug@example.com", "DebugPassword123!", []string{"user"})
		loginData := models.UserLoginBody{
			Email:    testUser.Email,
			Password: testUser.Password,
		}
		jsonData, err := json.Marshal(loginData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		t.Logf("Request URL: %s", server.URL+"/auth/login")
		t.Logf("Request Content-Type: application/json")
		t.Logf("Request Body: %s", string(jsonData))
		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		t.Logf("Response Status: %d", resp.StatusCode)
		t.Logf("Response Headers: %+v", resp.Header)
		t.Logf("Response Body: %s", string(body))
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestCreateSubforum_Integration(t *testing.T) {
	t.Run("CreateSubforumWithValidData", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "subforum_creator@example.com", "Password123!", []string{"user"})
		loginResp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		token := suite.ExtractTokenFromResponse(t, loginResp)
		createReq := models.SubforumCreateInput{
			Body: models.SubforumCreateBody{
				Slug:         "testslug1",
				Name:         "Test Subforum",
				Description:  "A subforum for testing.",
				SidebarText:  "Welcome!",
				RulesText:    "Be nice.",
				IsNSFW:       false,
				IsPrivate:    false,
				IsRestricted: false,
			},
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", "/subforums", token, createReq.Body)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 200 OK, got %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("CreateSubforumWithMissingFields", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "subforum_creator2@example.com", "Password123!", []string{"user"})
		loginResp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		token := suite.ExtractTokenFromResponse(t, loginResp)
		missingReq := models.SubforumCreateInput{
			Body: models.SubforumCreateBody{
				Name:        "No Slug",
				Description: "Missing slug field.",
			},
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", "/subforums", token, missingReq.Body)
		if resp.StatusCode != 400 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 400 Bad Request for missing slug, got %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("CreateSubforumWithoutCapability", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		noCapUser := suite.CreateTestUser(t, "nocap@example.com", "Password123!", []string{})
		loginResp := suite.LoginUser(t, server, noCapUser.Email, noCapUser.Password)
		noCapToken := suite.ExtractTokenFromResponse(t, loginResp)
		capabilityReq := models.SubforumCreateInput{
			Body: models.SubforumCreateBody{
				Slug:         "testslug2",
				Name:         "Test Subforum 2",
				Description:  "A subforum for capability testing.",
				SidebarText:  "Welcome!",
				RulesText:    "Be nice.",
				IsNSFW:       false,
				IsPrivate:    false,
				IsRestricted: false,
			},
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", "/subforums", noCapToken, capabilityReq.Body)
		if resp.StatusCode != 403 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 403 Forbidden for user without capability, got %d: %s", resp.StatusCode, string(body))
		}
	})
}

func TestPostWorkflow_Integration(t *testing.T) {
	t.Run("CompletePostWorkflow", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "post_creator@example.com", "Password123!", []string{"user"})
		testSubforum := suite.CreateTestSubforum(t, "test-subforum", "A test subforum", testUser.UserID, false)
		loginResp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		token := suite.ExtractTokenFromResponse(t, loginResp)
		postData := models.PostCreateBody{
			Title:     "Test Post",
			Content:   "This is a test post content.",
			PostType:  "text",
			URL:       "",
			IsNSFW:    false,
			IsSpoiler: false,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", "/subforums/"+testSubforum.Name+"/posts", token, postData)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 200 OK for post creation, got %d: %s", resp.StatusCode, string(body))
		}
		resp = suite.MakeAuthenticatedRequest(t, server, "GET", "/subforums/"+testSubforum.Name+"/posts", token, nil)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 200 OK for post retrieval, got %d: %s", resp.StatusCode, string(body))
		}
		var postsResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&postsResponse); err != nil {
			t.Fatalf("Failed to decode posts response: %v", err)
		}
		if posts, ok := postsResponse["posts"].([]interface{}); !ok || len(posts) == 0 {
			t.Error("Expected posts array in response")
		}
	})
}

func TestAuthRegistration_Integration(t *testing.T) {
	t.Run("RegisterUserWithValidData", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Test registration data using named struct
		registrationData := models.UserRegistrationBody{
			Email:       "newuser@example.com",
			Password:    "SecurePassword123!",
			DisplayName: "NewUser123",
		}

		jsonData, err := json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Read response body
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var response models.UserRegistrationResponseBody
		if err := json.Unmarshal(respBytes, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Debug: Print the actual response structure
		t.Logf("Registration response: %+v", response)

		// Verify response structure using the proper struct
		// Check user fields
		if response.UserID <= 0 {
			t.Error("Expected valid user_id in response")
		}
		if response.Email != "newuser@example.com" {
			t.Errorf("Expected email newuser@example.com, got %s", response.Email)
		}
		if len(response.Roles) == 0 {
			t.Error("Expected non-empty roles array")
		}
		if len(response.Capabilities) == 0 {
			t.Error("Expected non-empty capabilities array")
		}

		// Check pseudonym fields
		if response.PseudonymID == "" {
			t.Error("Expected non-empty pseudonym_id in response")
		}
		if response.DisplayName != "NewUser123" {
			t.Errorf("Expected display_name NewUser123, got %s", response.DisplayName)
		}
		if response.KarmaScore != 0 {
			t.Errorf("Expected karma_score 0, got %d", response.KarmaScore)
		}

		// Check token fields
		if response.AccessToken == "" {
			t.Error("Expected non-empty access_token in response")
		}
		if response.RefreshToken == "" {
			t.Error("Expected non-empty refresh_token in response")
		}
		if response.ExpiresIn <= 0 {
			t.Error("Expected positive expires_in value")
		}
	})

	t.Run("RegisterUserWithOptionalFields", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Test registration with optional fields using named struct
		registrationData := models.UserRegistrationBody{
			Email:       "userwithbio@example.com",
			Password:    "SecurePassword123!",
			DisplayName: "UserWithBio",
			Bio:         "This is my bio",
			WebsiteURL:  "https://example.com",
			Timezone:    "UTC",
			Language:    "en",
		}

		jsonData, err := json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify pseudonym was created correctly
		if body, ok := responseBody["body"].(map[string]interface{}); ok {
			if pseudonymID, ok := body["pseudonym_id"].(string); !ok || pseudonymID == "" {
				t.Error("Expected non-empty pseudonym_id in response")
			}
			if displayName, ok := body["display_name"].(string); !ok || displayName != "UserWithBio" {
				t.Errorf("Expected display_name UserWithBio, got %s", displayName)
			}
		}
	})

	t.Run("RegisterUserWithMissingRequiredFields", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Test missing email
		registrationData := models.UserRegistrationBody{
			Password:    "SecurePassword123!",
			DisplayName: "TestUser",
		}

		jsonData, err := json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 {
			t.Errorf("Expected error status for missing email, got %d", resp.StatusCode)
		}

		// Test missing password
		registrationData = models.UserRegistrationBody{
			Email:       "test@example.com",
			DisplayName: "TestUser",
		}

		jsonData, err = json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err = http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 {
			t.Errorf("Expected error status for missing password, got %d", resp.StatusCode)
		}

		// Test missing display_name
		registrationData = models.UserRegistrationBody{
			Email:    "test@example.com",
			Password: "SecurePassword123!",
		}

		jsonData, err = json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err = http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 {
			t.Errorf("Expected error status for missing display_name, got %d", resp.StatusCode)
		}
	})

	t.Run("RegisterUserWithDuplicateEmail", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create first user
		registrationData := models.UserRegistrationBody{
			Email:       "duplicate@example.com",
			Password:    "SecurePassword123!",
			DisplayName: "FirstUser",
		}

		jsonData, err := json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for first registration, got %d", resp.StatusCode)
		}

		// Try to register with same email
		registrationData.DisplayName = "SecondUser"
		jsonData, err = json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err = http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 400 {
			t.Errorf("Expected error status for duplicate email, got %d", resp.StatusCode)
		}
	})

	t.Run("RegisterUserAndVerifyPseudonymInDatabase", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Register user
		registrationData := models.UserRegistrationBody{
			Email:       "dbverify@example.com",
			Password:    "SecurePassword123!",
			DisplayName: "DBVerifyUser",
		}

		jsonData, err := json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Read response body
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var response models.UserRegistrationResponseBody
		if err := json.Unmarshal(respBytes, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Debug: Print the actual response structure
		t.Logf("Registration response: %+v", response)

		userID := int64(response.UserID)
		pseudonymID := response.PseudonymID

		// Verify pseudonym exists in database
		user, err := suite.UserDAO.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user from database: %v", err)
		}
		if user == nil {
			t.Fatal("User not found in database")
		}
		if user.Email != "dbverify@example.com" {
			t.Errorf("Expected email dbverify@example.com, got %s", user.Email)
		}

		// Verify pseudonym exists in database
		pseudonym, err := suite.SecurePseudonymDAO.GetPseudonymByID(context.Background(), pseudonymID)
		if err != nil {
			t.Fatalf("Failed to get pseudonym from database: %v", err)
		}
		if pseudonym == nil {
			t.Fatal("Pseudonym not found in database")
		}
		if pseudonym.DisplayName != "DBVerifyUser" {
			t.Errorf("Expected display_name DBVerifyUser, got %s", pseudonym.DisplayName)
		}
		if !pseudonym.IsActive.Valid || !pseudonym.IsActive.V {
			t.Error("Expected pseudonym to be active")
		}
	})

	t.Run("RegisterUserAndLoginWithCreatedPseudonym", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Register user
		registrationData := models.UserRegistrationBody{
			Email:       "loginverify@example.com",
			Password:    "SecurePassword123!",
			DisplayName: "LoginVerifyUser",
		}

		jsonData, err := json.Marshal(registrationData)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Login with the created user
		loginResp := suite.LoginUser(t, server, "loginverify@example.com", "SecurePassword123!")
		if loginResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for login, got %d", loginResp.StatusCode)
		}

		var loginResponseBody map[string]interface{}
		if err := json.NewDecoder(loginResp.Body).Decode(&loginResponseBody); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		// Verify login response contains pseudonym information
		if body, ok := loginResponseBody["body"].(map[string]interface{}); ok {
			if activePseudonymID, ok := body["active_pseudonym_id"].(string); !ok || activePseudonymID == "" {
				t.Error("Expected non-empty active_pseudonym_id in login response")
			}
			if displayName, ok := body["display_name"].(string); !ok || displayName != "LoginVerifyUser" {
				t.Errorf("Expected display_name LoginVerifyUser, got %s", displayName)
			}
			if pseudonyms, ok := body["pseudonyms"].([]interface{}); !ok || len(pseudonyms) == 0 {
				t.Error("Expected non-empty pseudonyms array in login response")
			} else {
				// Verify the pseudonym in the array
				if pseudonym, ok := pseudonyms[0].(map[string]interface{}); ok {
					if pid, ok := pseudonym["pseudonym_id"].(string); !ok || pid == "" {
						t.Error("Expected non-empty pseudonym_id in pseudonyms array")
					}
					if dn, ok := pseudonym["display_name"].(string); !ok || dn != "LoginVerifyUser" {
						t.Errorf("Expected display_name LoginVerifyUser in pseudonyms array, got %s", dn)
					}
				}
			}
		}
	})
}
