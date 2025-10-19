package main

import (
	"fmt"
	"strconv"
	"strings"
	"vsc_nft_mgmt/sdk"
)

const (
	maxNameLength = 25  // maxNameLength is the maximum length for collection or NFT names.
	maxDescLength = 100 // maxDescLength is the maximum length for collection or NFT descriptions.
)

// CreateCollection creates and saves a new collection.
//
//go:wasmexport col_create
func CreateCollection(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("input CSV is nil or empty")
	}

	parts := strings.Split(*payload, "|")
	if len(parts) != 2 {
		sdk.Abort("invalid CSV format: expected 2 fields (Name|Description)")
	}

	name := parts[0]
	description := parts[1]

	if name == "" {
		sdk.Abort("name is mandatory")
	}
	if len(name) > maxNameLength {
		sdk.Abort("name too long")
	}
	if len(description) > maxDescLength {
		sdk.Abort("description too long")
	}

	creator := sdk.GetEnvKey("msg.sender")
	collectionId := newCollectionID()

	saveCollection(collectionId, name, description, *creator)
	return nil
}

// GetCollection returns a collection by its ID.
//
//go:wasmexport col_get
func GetCollection(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("input CSV is nil or empty")
	}

	parts := strings.Split(*payload, "|")
	if len(parts) != 2 {
		sdk.Abort("invalid CSV format: expected 2 fields (owner|collectionID)")
	}

	collection := loadCollection(parts[0], parts[1])
	jsonStr := ToJSON(collection, "collection")
	return &jsonStr
}

// saveCollection persists a collection to state and emits a creation event.
func saveCollection(ID uint64, name string, description string, owner string) error {
	buf := make([]byte, 0, len(name)+len(description)+1)
	buf = append(buf, name...)
	buf = append(buf, '|')
	buf = append(buf, description...)
	// Save collection object.
	idKey := collectionKey(owner, strconv.FormatUint(ID, 10))
	sdk.StateSetObject(idKey, string(buf))

	// Emit creation event.
	EmitCollectionCreatedEvent(ID, owner)

	// Increment global collection counter.
	setCount(CollectionCount, ID+uint64(1))
	return nil
}

// loadCollection retrieves a collection from state by ID.
func loadCollection(owner string, id string) *string {
	key := collectionKey(owner, id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("collection %s:%s not found", owner, id))
	}
	return ptr
}

// collectionKey returns the state key for a collection ID.
func collectionKey(owner string, collectionId string) string {
	return owner + ":" + collectionId
}

// newCollectionID returns the next available collection ID from the counter.
func newCollectionID() uint64 {
	return getCount(CollectionCount)
}
