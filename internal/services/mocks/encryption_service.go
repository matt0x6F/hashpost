package mocks

import (
	"github.com/matt0x6f/hashpost/internal/services"
	"github.com/stretchr/testify/mock"
)

// MockEncryptionService is a mock of EncryptionServiceInterface
type MockEncryptionService struct {
	mock.Mock
}

// Ensure MockEncryptionService implements EncryptionServiceInterface
var _ services.EncryptionServiceInterface = (*MockEncryptionService)(nil)

// GenerateAESKey mocks base method
func (m *MockEncryptionService) GenerateAESKey() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// GenerateSignatureKeyPair mocks base method
func (m *MockEncryptionService) GenerateSignatureKeyPair() (*services.SignatureKeyPair, error) {
	args := m.Called()
	return args.Get(0).(*services.SignatureKeyPair), args.Error(1)
}

// EncryptAES mocks base method
func (m *MockEncryptionService) EncryptAES(plaintext, key []byte) (*services.EncryptedMessage, error) {
	args := m.Called(plaintext, key)
	return args.Get(0).(*services.EncryptedMessage), args.Error(1)
}

// DecryptAES mocks base method
func (m *MockEncryptionService) DecryptAES(key []byte, encryptedMsg *services.EncryptedMessage) ([]byte, error) {
	args := m.Called(key, encryptedMsg)
	return args.Get(0).([]byte), args.Error(1)
}

// EncryptMessageKey mocks base method
func (m *MockEncryptionService) EncryptMessageKey(messageKey, masterKey []byte) ([]byte, error) {
	args := m.Called(messageKey, masterKey)
	return args.Get(0).([]byte), args.Error(1)
}

// DecryptMessageKey mocks base method
func (m *MockEncryptionService) DecryptMessageKey(encryptedMessageKey, masterKey []byte) ([]byte, error) {
	args := m.Called(encryptedMessageKey, masterKey)
	return args.Get(0).([]byte), args.Error(1)
}

// SignMessage mocks base method
func (m *MockEncryptionService) SignMessage(message, privateKey []byte) ([]byte, error) {
	args := m.Called(message, privateKey)
	return args.Get(0).([]byte), args.Error(1)
}

// VerifySignature mocks base method
func (m *MockEncryptionService) VerifySignature(publicKey, message, signature []byte) bool {
	args := m.Called(publicKey, message, signature)
	return args.Bool(0)
}

// EncryptWithPublicKey mocks base method
func (m *MockEncryptionService) EncryptWithPublicKey(publicKeyPEM []byte, data []byte) ([]byte, error) {
	args := m.Called(publicKeyPEM, data)
	return args.Get(0).([]byte), args.Error(1)
}

// DecryptWithPrivateKey mocks base method
func (m *MockEncryptionService) DecryptWithPrivateKey(privateKeyPEM []byte, encryptedData []byte) ([]byte, error) {
	args := m.Called(privateKeyPEM, encryptedData)
	return args.Get(0).([]byte), args.Error(1)
}
