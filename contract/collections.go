package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// ============================
// Exported Collection Creation
// ============================

// CreateCollection creates a new NFT collection owned by the caller.
// Payload format: "<name>|<desc>|<metadata>"
// Name and description are validated for length. If either is invalid,
// the function aborts. This function is non-reversible by design as collections themselves are immutable.
//
//go:wasmexport col_create
func CreateCollection(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	parts := splitFixedPipe(*payload, 3)
	name := parts[0]
	desc := parts[1]
	meta := parts[2]

	// validation
	if len(name) == 0 || len(name) > maxNameLength {
		sdk.Abort("invalid name length")
	}
	if len(desc) > maxDescLength {
		sdk.Abort("description too long")
	}

	creator := *sdk.GetEnvKey("msg.sender")
	colNumber := getNextCollectionId(creator) // get collection count as next id

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

	idxKey := colIndexKey(creator, strconv.FormatUint(colNumber, 10))
	sdk.StateSetObject(idxKey, string(b))   // store collection
	updateUserCollectionCount(id, creator)  // increment collection counter for user
	EmitCollectionCreatedEvent(id, creator) // emit event for indexers

	return nil
}

// =============================
// Internal Collection Functions
// =============================

// loadCollection ensures that a given owner/collection pair exists.
// It returns the raw collection ID as a *string. This helper is used
// internally by minting and other operations that depend on collection status.
func loadCollection(ownerCollection string) *string {
	if ownerCollection == "" {
		sdk.Abort("empty collection id")
	}
	owner, col := splitOwnerCollection(ownerCollection)
	ptr := sdk.StateGetObject(colIndexKey(owner, col))
	if ptr == nil || *ptr == "" {
		sdk.Abort("collection not found")
	}
	return ptr
}

// colIndexKey returns the index "c_<owner>_<collection>" string for state lookups.
func colIndexKey(owner, col string) string {
	b := make([]byte, 0, 2+len(owner)+1+len(col))
	b = append(b, 'c', '_')
	b = append(b, owner...)
	b = append(b, '_')
	b = append(b, col...)
	return string(b)
}
