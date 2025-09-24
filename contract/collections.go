package main

import (
	"encoding/json"
	"fmt"
	"vsc_nft_mgmt/sdk"
)

const (
	maxNameLength = 100  // maximum length if the name attribute (used by collections and nfts)
	maxDescLength = 1000 // maximum length if the description attribute (used by collections and nfts)
)

type Collection struct {
	ID           uint64      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"desc"`
	Owner        sdk.Address `json:"owner"`
	CreationTxID string      `json:"txid"`
}

// function arguments
type CreateCollectionArgs struct {
	Name        string `json:"name"` // mandatory: name of the collection
	Description string `json:"desc"` // optional: description of the collection
}

//go:wasmexport col_create
func CreateCollection(payload *string) *string {
	// env := sdkInterface.GetEnv()
	input := FromJSON[CreateCollectionArgs](*payload, "collection args")

	input.Validate()
	env := sdk.GetEnv()
	creator := sdk.GetEnvKey("msg.sender")
	collectionId := newCollectionID()

	collection := Collection{
		ID:           collectionId,
		Owner:        sdk.Address(*creator),
		Name:         input.Name,
		Description:  input.Description,
		CreationTxID: env.TxId,
	}
	saveCollection(&collection)
	return nil
}

// GET FUNCTIONS
// returns an collection for a given collection id

//go:wasmexport col_get
func GetCollection(id *string) *string {
	collection := loadCollection(StringToUInt64(id))
	jsonStr := ToJSON(collection, "collection")
	return &jsonStr
}

// Contract State Persistence
func saveCollection(collection *Collection) error {
	b, err := json.Marshal(collection)
	if err != nil {
		sdk.Abort("failed to marshal collection")
	}

	// save collection itself
	idKey := collectionKey(collection.ID)
	sdk.StateSetObject(idKey, string(b))
	EmitCollectionCreatedEvent(collection.ID, collection.Owner.String())
	// increase global collection counter
	setCount(CollectionCount, collection.ID+uint64(1))
	return nil
}

func loadCollection(id uint64) *Collection {
	key := collectionKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("collection %d not found", id))
	}
	collection := FromJSON[Collection](*ptr, "collection")
	return collection
}

func (c *CreateCollectionArgs) Validate() {
	if c.Name == "" {
		sdk.Abort("name is mandatory")
	}
	if len(c.Name) > maxNameLength {
		sdk.Abort(fmt.Sprintf("name: max %d chars", maxNameLength))
	}
	if len(c.Description) > maxDescLength {
		sdk.Abort(fmt.Sprintf("desc: max %d chars", maxDescLength))
	}
}

func collectionKey(collectionId uint64) string {
	return fmt.Sprintf("c:%d", collectionId)
}

func newCollectionID() uint64 {
	return getCount(CollectionCount)
}
