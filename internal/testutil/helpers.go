package testutil

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// StringPtr returns a pointer to a string (exported version)
func StringPtr(s string) *string {
	return &s
}
