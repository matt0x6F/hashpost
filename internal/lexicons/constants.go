package lexicons

// This file is deprecated - use internal/lexicons/generated/types.go instead
// Keeping for backward compatibility during transition

// HashPost Collection Constants
const (
	// Feed Collections
	CollectionFeedPost     = "com.hashpost.feed.post"
	CollectionFeedSubforum = "com.hashpost.feed.subforum"
	CollectionFeedComment  = "com.hashpost.feed.comment"

	// Record Types
	RecordTypePost     = "com.hashpost.feed.post"
	RecordTypeSubforum = "com.hashpost.feed.subforum"
	RecordTypeComment  = "com.hashpost.feed.comment"
)

// HashPost Field Constants
const (
	// Common fields
	FieldText        = "text"
	FieldCreatedAt   = "createdAt"
	FieldName        = "name"
	FieldSlug        = "slug"
	FieldDescription = "description"

	// Post-specific fields
	FieldTitle   = "title"
	FieldContent = "content"

	// Comment-specific fields
	FieldPost   = "post"
	FieldParent = "parent"

	// Subforum-specific fields
	FieldCreatedBy  = "createdBy"
	FieldPrefixType = "prefixType"
)

// HashPost URI Constants
const (
	// URI prefixes
	URIPrefixPost     = "at://"
	URIPrefixSubforum = "at://"

	// Collection paths
	PathFeedPost     = "com.hashpost.feed.post"
	PathFeedSubforum = "com.hashpost.feed.subforum"
)
