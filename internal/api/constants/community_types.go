package constants

// Community Types
const (
	CommunityTypeTopical    = "t" // t/ - Topical (generic subjects like t/programming, t/cooking)
	CommunityTypeGeographic = "g" // g/ - Geographic (location-based like g/seattle, g/tacoma)
	CommunityTypeBranded    = "b" // b/ - Branded (company/brand-owned like b/apple, b/minecraft)
	CommunityTypeCreator    = "c" // c/ - Creator (individual creator-owned like c/joerogan)
)

// Governance Styles
const (
	GovernanceStyleDemocratic = "democratic" // Owner commits to elected moderators/community input
	GovernanceStyleOwned      = "owned"      // Owner directly manages all moderators
)

// CommunityTypeInfo provides metadata about each community type
type CommunityTypeInfo struct {
	Type         string
	Prefix       string
	Name         string
	Description  string
	Governance   string // Default governance style
	CanElectMods bool   // Whether democratic communities can elect moderators
}

// CommunityTypes maps type codes to their metadata
var CommunityTypes = map[string]CommunityTypeInfo{
	CommunityTypeTopical: {
		Type:         CommunityTypeTopical,
		Prefix:       "t/",
		Name:         "Topical",
		Description:  "Generic subjects and topics",
		Governance:   GovernanceStyleDemocratic,
		CanElectMods: true,
	},
	CommunityTypeGeographic: {
		Type:         CommunityTypeGeographic,
		Prefix:       "g/",
		Name:         "Geographic",
		Description:  "Location-based communities",
		Governance:   GovernanceStyleDemocratic,
		CanElectMods: true,
	},
	CommunityTypeBranded: {
		Type:         CommunityTypeBranded,
		Prefix:       "b/",
		Name:         "Branded",
		Description:  "Company and brand-owned communities",
		Governance:   GovernanceStyleOwned,
		CanElectMods: false,
	},
	CommunityTypeCreator: {
		Type:         CommunityTypeCreator,
		Prefix:       "c/",
		Name:         "Creator",
		Description:  "Individual creator-owned communities",
		Governance:   GovernanceStyleOwned,
		CanElectMods: false,
	},
}

// GetCommunityTypeInfo returns metadata for a community type
func GetCommunityTypeInfo(communityType string) (CommunityTypeInfo, bool) {
	info, exists := CommunityTypes[communityType]
	return info, exists
}

// IsValidCommunityType checks if a community type is valid
func IsValidCommunityType(communityType string) bool {
	_, exists := CommunityTypes[communityType]
	return exists
}

// IsDemocraticCommunity checks if a community type uses democratic governance
func IsDemocraticCommunity(communityType string) bool {
	info, exists := CommunityTypes[communityType]
	return exists && info.Governance == GovernanceStyleDemocratic
}

// IsOwnedCommunity checks if a community type uses owned governance
func IsOwnedCommunity(communityType string) bool {
	info, exists := CommunityTypes[communityType]
	return exists && info.Governance == GovernanceStyleOwned
}

// GetCommunityURL returns the full URL for a community
func GetCommunityURL(communityType, name string) string {
	info, exists := CommunityTypes[communityType]
	if !exists {
		return ""
	}
	return info.Prefix + name
}
