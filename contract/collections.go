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
	sdk.Log("collection start")
	if payload == nil || *payload == "" {
		sdk.Abort("input CSV is nil or empty")
	}

	parts := strings.Split(*payload, "|")
	if len(parts) != 2 {
		sdk.Abort("invalid CSV format: expected 2 fields (Name|Description)")
	}

	sdk.Log("collection args parsed")

	name := parts[0]
	description := parts[1]
	sdk.Log(name)
	sdk.Log(description)

	if name == "" {
		sdk.Abort("name is mandatory")
	}
	if len(name) > maxNameLength {
		sdk.Abort("name too long")
	}
	if len(description) > maxDescLength {
		sdk.Abort("description too long")
	}
	sdk.Log("collection args validated")

	creator := sdk.GetEnvKey("msg.sender")
	sdk.Log(*creator)
	collectionId := getCount(CollectionCount)
	sdk.Log(fmt.Sprintf("%d", collectionId))

	saveCollection(collectionId, name, description, *creator)
	sdk.Log("collection stored")
	return nil
}

// GetCollection returns a collection by its ID.
//
//go:wasmexport col_get
func GetCollection(payload *string) *string {
	collection := loadCollection(*payload)
	jsonStr := ToJSON(collection, "collection")
	return &jsonStr
}

// saveCollection persists a collection to state and emits a creation event.
func saveCollection(ID uint64, name string, description string, owner string) error {
	// ! continue here
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
func loadCollection(ownerCollection string) *string {
	ptr := sdk.StateGetObject(ownerCollection)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("%s not found", ownerCollection))
	}
	return ptr
}

// collectionKey returns the state key for a collection ID.
func collectionKey(owner string, collectionId string) string {
	return owner + "/" + collectionId
}
