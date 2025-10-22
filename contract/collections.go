package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// kColCore is the storage prefix for the core collection record.
// The underlying value follows the format: tx|name|desc|meta.
const (
	kColCore byte = 0x10
)

// colIndexKey returns the ASCII index key used to map a human-readable
// "<owner>_<collection>" string to a numeric collection ID.
// This function is critical for security, as it enforces owner↔collection binding.
func colIndexKey(owner, col string) string {
	b := make([]byte, 0, 2+len(owner)+1+len(col))
	b = append(b, 'c', '_')
	b = append(b, owner...)
	b = append(b, '_')
	b = append(b, col...)
	return string(b)
}

// colCoreKey builds the binary key for storing immutable collection metadata.
// The resulting key uses a single-byte prefix followed by the little-endian ID value.
func colCoreKey(id uint64) string {
	b := make([]byte, 0, 1+8)
	b = append(b, kColCore)
	b = packU64LE(id, b)
	return string(b)
}

// =============================
// Exported ABI: string-only
// =============================

// CreateCollection creates a new NFT collection owned by the caller.
// Payload format: "<name>|<desc>|<metadata>"
// Name and description are validated for length. If either is invalid,
// the function aborts. This function is non-reversible by design (immutabel).
//
//go:wasmexport col_create
func CreateCollection(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	p := *payload
	parts := splitFixedPipe(p, 3)
	name := parts[0]
	desc := parts[1]
	meta := parts[2]

	if len(name) == 0 || len(name) > maxNameLength {
		sdk.Abort("invalid name length")
	}
	if len(desc) > maxDescLength {
		sdk.Abort("description too long")
	}

	owner := *sdk.GetEnvKey("msg.sender")
	colNumber := getCount(CollectionCount)
	col := strconv.FormatUint(colNumber, 10)

	idxKey := colIndexKey(owner, col)
	if ptr := sdk.StateGetObject(idxKey); ptr != nil && *ptr != "" {
		sdk.Abort("collection exists")
	}

	id := colNumber
	txID := sdk.GetEnvKey("tx.id")

	b := make([]byte, 0, len(*txID)+1+len(name)+1+len(desc)+1+len(meta))
	b = append(b, (*txID)...)
	b = append(b, '|')
	b = append(b, name...)
	b = append(b, '|')
	b = append(b, desc...)
	b = append(b, '|')
	b = append(b, meta...)

	sdk.StateSetObject(colCoreKey(id), string(b))

	// More efficient than FormatUint
	tmp := make([]byte, 0, 20)
	tmp = strconv.AppendUint(tmp, id, 10)
	sdk.StateSetObject(idxKey, string(tmp))

	EmitCollectionCreatedEvent(id, owner)
	setCount(CollectionCount, id+1)
	return nil
}

// GetCollection retrieves metadata for a specific collection.
// Payload format: "<owner>_<collection>".
// The owner data is validated via the collection index, so spoofing
// a wrong owner in the input will be rejected.
//
//go:wasmexport col_get
func GetCollection(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty id")
	}
	ownerCol := *payload
	owner, col := splitOwnerCollection(ownerCol, "collectionLookup")

	// Validate existence & resolve numeric collection ID
	colIDPtr := sdk.StateGetObject(colIndexKey(owner, col))
	if colIDPtr == nil || *colIDPtr == "" {
		sdk.Abort("collection not found")
	}
	colID := mustParseUint64(*colIDPtr)

	// Load core collection data
	corePtr := sdk.StateGetObject(colCoreKey(colID))
	if corePtr == nil || *corePtr == "" {
		sdk.Abort("collection core missing")
	}
	tx, name, desc, meta := parse4(*corePtr)

	// Return compact JSON response
	b := make([]byte, 0, 64+len(owner)+len(name)+len(desc)+len(meta)+len(tx))
	b = append(b, '{', '"', 'i', 'd', '"', ':')
	b = strconv.AppendUint(b, colID, 10)
	b = append(b, ',', '"', 'o', 'w', 'n', 'e', 'r', '"', ':', '"')
	b = append(b, owner...)
	b = append(b, '"', ',', '"', 'n', 'a', 'm', 'e', '"', ':', '"')
	b = append(b, name...)
	b = append(b, '"', ',', '"', 'd', 'e', 's', 'c', '"', ':', '"')
	b = append(b, desc...)
	b = append(b, '"', ',', '"', 'm', 'e', 't', 'a', '"', ':', '"')
	b = append(b, meta...)
	b = append(b, '"', ',', '"', 't', 'x', '"', ':', '"')
	b = append(b, tx...)
	b = append(b, '"', '}')

	json := string(b)
	return &json
}

// loadCollection ensures that a given owner/collection pair exists.
// It returns the raw collection ID as a *string. This helper is used
// internally by minting and other operations that depend on collection status.
func loadCollection(ownerCollection string) *string {
	if ownerCollection == "" {
		sdk.Abort("empty collection id")
	}
	owner, col := splitOwnerCollection(ownerCollection, "collectionLookup")
	ptr := sdk.StateGetObject(colIndexKey(owner, col))
	if ptr == nil || *ptr == "" {
		sdk.Abort("collection not found")
	}
	return ptr // the numeric collection ID as string
}
