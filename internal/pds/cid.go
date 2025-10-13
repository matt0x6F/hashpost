package pds

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/bluesky-social/indigo/mst"
	"github.com/ipfs/go-cid"
	cbornode "github.com/ipfs/go-ipld-cbor"
	"github.com/multiformats/go-multihash"
)

// CIDService handles content-addressed identifier computation for atproto records
type CIDService struct {
	logger *slog.Logger
}

// NewCIDService creates a new CID service
func NewCIDService(logger *slog.Logger) *CIDService {
	return &CIDService{
		logger: logger,
	}
}

// ComputeRecordCID computes a proper content-addressed CID for an atproto record
func (cs *CIDService) ComputeRecordCID(ctx context.Context, record map[string]interface{}) (string, error) {
	// Serialize the record to CBOR format for proper content addressing
	cborData, err := cs.serializeRecordToCBOR(record)
	if err != nil {
		return "", fmt.Errorf("failed to serialize record to CBOR: %w", err)
	}

	// Create a proper CID using the CBOR data
	cid, err := cs.createCIDFromData(cborData)
	if err != nil {
		return "", fmt.Errorf("failed to create CID: %w", err)
	}

	cs.logger.Debug("Computed CID for record", "cid", cid.String(), "size", len(cborData))
	return cid.String(), nil
}

// ComputeRepoCID computes a CID for a repository using MST
func (cs *CIDService) ComputeRepoCID(ctx context.Context, repoData []byte) (string, error) {
	// Create a proper CID using the repository data
	cid, err := cs.createCIDFromData(repoData)
	if err != nil {
		return "", fmt.Errorf("failed to create CID for repository: %w", err)
	}

	cs.logger.Debug("Computed CID for repository", "cid", cid.String(), "size", len(repoData))
	return cid.String(), nil
}

// ValidateCID validates that a CID string is properly formatted
func (cs *CIDService) ValidateCID(cidStr string) error {
	// Parse the CID to validate its format
	_, err := cid.Decode(cidStr)
	if err != nil {
		return fmt.Errorf("invalid CID format: %w", err)
	}
	return nil
}

// serializeRecordToCBOR serializes a record to CBOR format using proper CBOR encoding
func (cs *CIDService) serializeRecordToCBOR(record map[string]interface{}) ([]byte, error) {
	// Convert the record to a format suitable for CBOR encoding
	// Create a proper IPLD node from the record
	node, err := cbornode.WrapObject(record, multihash.SHA2_256, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap record in CBOR: %w", err)
	}

	// Get the raw CBOR data
	cborData := node.RawData()

	return cborData, nil
}

// createCIDFromData creates a proper CID from data using SHA256 multihash
func (cs *CIDService) createCIDFromData(data []byte) (cid.Cid, error) {
	// Create SHA256 multihash
	hash, err := multihash.Sum(data, multihash.SHA2_256, -1)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("failed to create multihash: %w", err)
	}

	// Create CID with raw codec (0x55) and SHA256 hash
	c := cid.NewCidV1(cid.Raw, hash)

	return c, nil
}

// ComputeMSTNodeCID computes a CID for an MST node using proper CID creation
func (cs *CIDService) ComputeMSTNodeCID(ctx context.Context, nodeData []byte) (string, error) {
	// Create a proper CID using the node data
	c, err := cs.createCIDFromData(nodeData)
	if err != nil {
		return "", fmt.Errorf("failed to create CID for MST node: %w", err)
	}

	cs.logger.Debug("Computed CID for MST node", "cid", c.String(), "size", len(nodeData))
	return c.String(), nil
}

// ComputeBlobCID computes a CID for a blob (binary data) using proper CID creation
func (cs *CIDService) ComputeBlobCID(data []byte) (string, error) {
	// Create a proper CID using the blob data
	c, err := cs.createCIDFromData(data)
	if err != nil {
		return "", fmt.Errorf("failed to create CID for blob: %w", err)
	}

	cs.logger.Debug("Computed CID for blob", "cid", c.String(), "size", len(data))
	return c.String(), nil
}

// GetCIDInfo returns information about a CID
func (cs *CIDService) GetCIDInfo(cidStr string) (*CIDInfo, error) {
	// Parse the CID to get proper information
	c, err := cid.Decode(cidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CID: %w", err)
	}

	info := &CIDInfo{
		CID:     cidStr,
		Version: uint64(c.Version()),
		Codec:   uint64(c.Type()),
		Hash:    c.Hash().HexString(),
	}

	return info, nil
}

// CIDInfo contains information about a CID
type CIDInfo struct {
	CID     string `json:"cid"`
	Version uint64 `json:"version"`
	Codec   uint64 `json:"codec"`
	Hash    string `json:"hash"`
}

// ValidateAtprotoURI validates an atproto URI format
func (cs *CIDService) ValidateAtprotoURI(uri string) error {
	// Basic validation - check if it starts with "at://"
	if len(uri) < 6 || !bytes.HasPrefix([]byte(uri), []byte("at://")) {
		return fmt.Errorf("invalid atproto URI format: %s", uri)
	}
	return nil
}

// ValidateAtprotoDID validates an atproto DID format
func (cs *CIDService) ValidateAtprotoDID(did string) error {
	// Basic validation - check if it starts with "did:"
	if len(did) < 4 || !bytes.HasPrefix([]byte(did), []byte("did:")) {
		return fmt.Errorf("invalid atproto DID format: %s", did)
	}
	return nil
}

// CreateMSTForRepo creates a new MST for a repository
func (cs *CIDService) CreateMSTForRepo(ctx context.Context) (*mst.MerkleSearchTree, error) {
	// Create a new MST instance
	tree := mst.NewEmptyMST(nil) // We'll need to provide a proper blockstore
	return tree, nil
}

// AddRecordToMST adds a record to an MST and returns the new root CID
func (cs *CIDService) AddRecordToMST(ctx context.Context, tree *mst.MerkleSearchTree, key string, record map[string]interface{}) (string, error) {
	// Serialize the record to CBOR and create a CID
	cborData, err := cs.serializeRecordToCBOR(record)
	if err != nil {
		return "", fmt.Errorf("failed to serialize record: %w", err)
	}

	// Create a CID for the record data
	recordCid, err := cs.createCIDFromData(cborData)
	if err != nil {
		return "", fmt.Errorf("failed to create CID for record: %w", err)
	}

	// Add the record to the MST (returns new tree and error)
	newTree, err := tree.Add(ctx, key, recordCid, 0) // 0 for knownZeros
	if err != nil {
		return "", fmt.Errorf("failed to add record to MST: %w", err)
	}

	// Get the root CID of the new MST
	rootCid, err := newTree.GetPointer(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get root CID: %w", err)
	}

	return rootCid.String(), nil
}

// GetRecordFromMST retrieves a record from an MST
func (cs *CIDService) GetRecordFromMST(ctx context.Context, tree *mst.MerkleSearchTree, key string) (cid.Cid, error) {
	// Get the record CID from the MST
	recordCid, err := tree.Get(ctx, key)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("failed to get record from MST: %w", err)
	}

	return recordCid, nil
}

// RemoveRecordFromMST removes a record from an MST and returns the new root CID
func (cs *CIDService) RemoveRecordFromMST(ctx context.Context, tree *mst.MerkleSearchTree, key string) (string, error) {
	// Remove the record from the MST (returns new tree and error)
	newTree, err := tree.Delete(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to remove record from MST: %w", err)
	}

	// Get the root CID of the new MST
	rootCid, err := newTree.GetPointer(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get root CID: %w", err)
	}

	return rootCid.String(), nil
}
