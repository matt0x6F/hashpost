package ibe

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockIBESystem is a mock implementation of IBESystemInterface
type MockIBESystem struct {
	mock.Mock
	keyVersion int32
}

// NewMockIBESystem creates a new mock IBE system
func NewMockIBESystem() *MockIBESystem {
	return &MockIBESystem{
		keyVersion: 1,
	}
}

func (m *MockIBESystem) GetDomainMasters() map[string][]byte {
	args := m.Called()
	return args.Get(0).(map[string][]byte)
}

func (m *MockIBESystem) GetKeyVersion() int32 {
	args := m.Called()
	if len(args) > 0 {
		return args.Get(0).(int32)
	}
	return m.keyVersion
}

func (m *MockIBESystem) SetKeyVersion(version int32) {
	m.Called(version)
	m.keyVersion = version
}

func (m *MockIBESystem) DecryptIdentityWithVersion(encryptedData []byte, domain string, keyVersion int32) (string, string, error) {
	args := m.Called(encryptedData, domain, keyVersion)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockIBESystem) EncryptFingerprintMapping(fingerprint, pseudonymID, domain string, keyVersion int32) ([]byte, error) {
	args := m.Called(fingerprint, pseudonymID, domain, keyVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIBESystem) GeneratePseudonym(userID int64, context string, version int32) string {
	args := m.Called(userID, context, version)
	return args.String(0)
}

func (m *MockIBESystem) GenerateCorrelationKeyForVersion(role, scope string, timeWindow time.Duration, version int32) []byte {
	args := m.Called(role, scope, timeWindow, version)
	return args.Get(0).([]byte)
}

func (m *MockIBESystem) AddKeyVersion(version int32, domainKeys map[string][]byte, salt string) {
	m.Called(version, domainKeys, salt)
}

func (m *MockIBESystem) DeprecateKeyVersion(version int32) error {
	args := m.Called(version)
	return args.Error(0)
}
