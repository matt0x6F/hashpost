package ibe

import "time"

// IBESystemInterface defines the interface for IBE system operations
// This interface enables better testability and dependency injection
type IBESystemInterface interface {
	GetDomainMasters() map[string][]byte
	GetKeyVersion() int32
	SetKeyVersion(version int32)
	DecryptIdentityWithVersion(encryptedData []byte, domain string, keyVersion int32) (string, string, error)
	DecryptIdentity(encryptedMapping []byte, adminKey []byte) (string, string, error)
	EncryptFingerprintMapping(fingerprint, pseudonymID, domain string, keyVersion int32) ([]byte, error)
	GeneratePseudonym(userID int64, context string, version int32) string
	GenerateCorrelationKeyForVersion(role, scope string, timeWindow time.Duration, version int32) []byte
	GenerateRoleKey(role string, scope string, expiration time.Time) []byte
	AddKeyVersion(version int32, domainKeys map[string][]byte, salt string)
	DeprecateKeyVersion(version int32) error
}
