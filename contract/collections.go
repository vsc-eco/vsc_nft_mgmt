package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// =============================
// Collection binary key prefixes
// =============================
const (
	kColCore byte = 0x10 // id -> "tx|name|desc|meta"
)

// ASCII index: "<owner>_<collection>" → numeric ID (as string)
func colIndexKey(owner, col string) string { return "c_" + owner + "_" + col }

// =============================
// Binary key builder
// =============================
func colCoreKey(id uint64) string {
	b := make([]byte, 0, 1+8)
	b = append(b, kColCore)
	b = packU64LE(id, b)
	return string(b)
}

// =============================
// Exported ABI: string-only
// =============================

//go:wasmexport col_create
func CreateCollection(payload *string) *string {
	// Payload: "<name>|<desc>|<metadata>"
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

	// Ensure not already exists
	idxKey := colIndexKey(owner, col)
	if ptr := sdk.StateGetObject(idxKey); ptr != nil && *ptr != "" {
		sdk.Abort("collection exists")
	}

	id := getCount(CollectionCount)
	txID := sdk.GetEnvKey("tx.id")

	// Core format: tx|name|desc|meta
	b := make([]byte, 0, len(*txID)+1+len(name)+1+len(desc)+1+len(meta))
	b = append(b, (*txID)...)
	b = append(b, '|')
	b = append(b, name...)
	b = append(b, '|')
	b = append(b, desc...)
	b = append(b, '|')
	b = append(b, meta...)

	sdk.StateSetObject(colCoreKey(id), string(b))
	sdk.StateSetObject(idxKey, strconv.FormatUint(id, 10))

	EmitCollectionCreatedEvent(id, owner)
	setCount(CollectionCount, id+1)
	return nil
}

//go:wasmexport col_get
func GetCollection(payload *string) *string {
	// Payload: "<owner>_<collection>"
	if payload == nil || *payload == "" {
		sdk.Abort("empty id")
	}
	ownerCol := *payload
	owner, col := splitOwnerCollection(ownerCol, "collectionLookup")

	// Validate via index
	colIDPtr := sdk.StateGetObject(colIndexKey(owner, col))
	if colIDPtr == nil || *colIDPtr == "" {
		sdk.Abort("collection not found")
	}
	colID := mustParseUint64(*colIDPtr)

	// Load core
	corePtr := sdk.StateGetObject(colCoreKey(colID))
	if corePtr == nil || *corePtr == "" {
		sdk.Abort("collection core missing")
	}
	tx, name, desc, meta := parse4(*corePtr)

	// JSON: {"id":N,"owner":"...","name":"...","desc":"...","meta":"...","tx":"..."}
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

// =============================
// Helpers
// =============================
// loadCollection validates that a given "<owner>_<collection>" exists
// and returns the collection ID as a *string.
func loadCollection(ownerCollection string) *string {
	if ownerCollection == "" {
		sdk.Abort("empty collection id")
	}
	owner, col := splitOwnerCollection(ownerCollection, "collectionLookup")
	ptr := sdk.StateGetObject(colIndexKey(owner, col))
	if ptr == nil || *ptr == "" {
		sdk.Abort("collection not found")
	}
	return ptr // This is the numeric collection ID as a string
}
