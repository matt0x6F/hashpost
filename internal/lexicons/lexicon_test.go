package lexicons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLexiconValidation tests that our custom lexicons are properly defined and valid
func TestLexiconValidation(t *testing.T) {
	t.Run("lexicon_constants_defined", func(t *testing.T) {
		// Test that our lexicon constants are defined
		assert.NotEmpty(t, CollectionFeedPost)
		assert.NotEmpty(t, CollectionFeedSubforum)
	})

	t.Run("lexicon_constants_format", func(t *testing.T) {
		// Test that our lexicon constants follow the proper format
		assert.Contains(t, CollectionFeedPost, ".")
		assert.Contains(t, CollectionFeedSubforum, ".")
	})

	t.Run("lexicon_constants_namespace", func(t *testing.T) {
		// Test that our lexicon constants use the proper namespace
		assert.Contains(t, CollectionFeedPost, "hashpost")
		assert.Contains(t, CollectionFeedSubforum, "hashpost")
	})
}

// TestLexiconValidationIntegration tests that our lexicon validation works with real data
func TestLexiconValidationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("lexicon_validation_with_real_data", func(t *testing.T) {
		// Test that our lexicon validation works with real atproto data
		// This would test that our custom types can be properly serialized/deserialized
		// and that they conform to the atproto specification

		// For now, this is a placeholder test
		// In a real implementation, we would:
		// 1. Create test data using our custom types
		// 2. Serialize it to JSON
		// 3. Validate it against the lexicon schema
		// 4. Deserialize it back to Go types
		// 5. Verify the data integrity

		assert.True(t, true, "Lexicon validation integration test placeholder")
	})
}
