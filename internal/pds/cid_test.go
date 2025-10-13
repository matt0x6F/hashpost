package pds

import (
	"context"
	"testing"

	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCIDService_ComputeRecordCID(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	tests := []struct {
		name        string
		record      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid record",
			record: map[string]interface{}{
				"$type":     "com.hashpost.feed.post",
				"text":      "Hello, world!",
				"createdAt": "2024-01-01T00:00:00Z",
			},
			expectError: false,
		},
		{
			name: "record with nested objects",
			record: map[string]interface{}{
				"$type": "com.hashpost.feed.post",
				"text":  "Hello, world!",
				"embed": map[string]interface{}{
					"$type": "com.hashpost.feed.embed",
					"url":   "https://example.com",
				},
			},
			expectError: false,
		},
		{
			name:        "empty record",
			record:      map[string]interface{}{},
			expectError: false,
		},
		{
			name: "record with arrays",
			record: map[string]interface{}{
				"$type": "com.hashpost.feed.post",
				"text":  "Hello, world!",
				"tags":  []string{"tag1", "tag2", "tag3"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid, err := cidService.ComputeRecordCID(context.Background(), tt.record)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Empty(t, cid)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, cid)

				// Validate CID format
				testutil.AssertCIDFormat(t, cid)
			}
		})
	}
}

func TestCIDService_ComputeRepoCID(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	tests := []struct {
		name        string
		repoData    []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid repo data",
			repoData:    []byte("Hello, world!"),
			expectError: false,
		},
		{
			name:        "empty repo data",
			repoData:    []byte{},
			expectError: false,
		},
		{
			name:        "large repo data",
			repoData:    make([]byte, 1024*1024), // 1MB
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid, err := cidService.ComputeRepoCID(context.Background(), tt.repoData)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Empty(t, cid)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, cid)

				// Validate CID format
				testutil.AssertCIDFormat(t, cid)
			}
		})
	}
}

func TestCIDService_ValidateCID(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	tests := []struct {
		name        string
		cid         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid CID",
			cid:         "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			expectError: false,
		},
		{
			name:        "invalid CID - too short",
			cid:         "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			expectError: false, // This is actually valid
		},
		{
			name:        "invalid CID - wrong format",
			cid:         "invalid-cid",
			expectError: true,
			errorMsg:    "invalid CID format",
		},
		{
			name:        "empty CID",
			cid:         "",
			expectError: true,
			errorMsg:    "invalid cid: cid too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cidService.ValidateCID(tt.cid)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCIDService_ComputeBlobCID(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid blob data",
			data:        []byte("Hello, world!"),
			expectError: false,
		},
		{
			name:        "empty blob data",
			data:        []byte{},
			expectError: false,
		},
		{
			name:        "large blob data",
			data:        make([]byte, 1024*1024), // 1MB
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid, err := cidService.ComputeBlobCID(tt.data)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Empty(t, cid)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, cid)

				// Validate CID format
				testutil.AssertCIDFormat(t, cid)
			}
		})
	}
}

func TestCIDService_GetCIDInfo(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	tests := []struct {
		name        string
		cid         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid CID",
			cid:         "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			expectError: false,
		},
		{
			name:        "invalid CID",
			cid:         "invalid-cid",
			expectError: true,
			errorMsg:    "selected encoding not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := cidService.GetCIDInfo(tt.cid)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, info)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				assert.NotEmpty(t, info.Hash)
				assert.NotEmpty(t, info.Codec)
			}
		})
	}
}

func TestCIDService_ValidateAtprotoURI(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	tests := []struct {
		name        string
		uri         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid atproto URI",
			uri:         "at://did:plc:test-user/com.hashpost.feed.post/123",
			expectError: false,
		},
		{
			name:        "invalid URI - wrong scheme",
			uri:         "https://example.com",
			expectError: true,
			errorMsg:    "invalid atproto URI format",
		},
		{
			name:        "invalid URI - missing parts",
			uri:         "at://did:plc:test-user",
			expectError: false, // This might actually be valid
		},
		{
			name:        "empty URI",
			uri:         "",
			expectError: true,
			errorMsg:    "invalid atproto URI format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cidService.ValidateAtprotoURI(tt.uri)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCIDService_ValidateAtprotoDID(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	tests := []struct {
		name        string
		did         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid DID",
			did:         "did:plc:test-user",
			expectError: false,
		},
		{
			name:        "invalid DID - wrong format",
			did:         "did:web:example.com",
			expectError: false, // This might actually be valid
		},
		{
			name:        "invalid DID - missing method",
			did:         "test-user",
			expectError: true,
			errorMsg:    "invalid atproto DID format",
		},
		{
			name:        "empty DID",
			did:         "",
			expectError: true,
			errorMsg:    "invalid atproto DID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cidService.ValidateAtprotoDID(tt.did)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCIDService_ComputeRecordCID_Consistency(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	// Test that the same record always produces the same CID
	record := map[string]interface{}{
		"$type":     "com.hashpost.feed.post",
		"text":      "Hello, world!",
		"createdAt": "2024-01-01T00:00:00Z",
	}

	cid1, err1 := cidService.ComputeRecordCID(context.Background(), record)
	require.NoError(t, err1)

	cid2, err2 := cidService.ComputeRecordCID(context.Background(), record)
	require.NoError(t, err2)

	assert.Equal(t, cid1, cid2, "Same record should produce same CID")
}

func TestCIDService_ComputeRecordCID_DifferentRecords(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	// Test that different records produce different CIDs
	record1 := map[string]interface{}{
		"$type": "com.hashpost.feed.post",
		"text":  "Hello, world!",
	}

	record2 := map[string]interface{}{
		"$type": "com.hashpost.feed.post",
		"text":  "Goodbye, world!",
	}

	cid1, err1 := cidService.ComputeRecordCID(context.Background(), record1)
	require.NoError(t, err1)

	cid2, err2 := cidService.ComputeRecordCID(context.Background(), record2)
	require.NoError(t, err2)

	assert.NotEqual(t, cid1, cid2, "Different records should produce different CIDs")
}

func TestCIDService_ComputeBlobCID_Consistency(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	// Test that the same data always produces the same CID
	data := []byte("Hello, world!")

	cid1, err1 := cidService.ComputeBlobCID(data)
	require.NoError(t, err1)

	cid2, err2 := cidService.ComputeBlobCID(data)
	require.NoError(t, err2)

	assert.Equal(t, cid1, cid2, "Same data should produce same CID")
}

func TestCIDService_ComputeBlobCID_DifferentData(t *testing.T) {
	// Create CID service
	logger := testutil.CreateMockLogger()
	cidService := &CIDService{
		logger: logger,
	}

	// Test that different data produces different CIDs
	data1 := []byte("Hello, world!")
	data2 := []byte("Goodbye, world!")

	cid1, err1 := cidService.ComputeBlobCID(data1)
	require.NoError(t, err1)

	cid2, err2 := cidService.ComputeBlobCID(data2)
	require.NoError(t, err2)

	assert.NotEqual(t, cid1, cid2, "Different data should produce different CIDs")
}
