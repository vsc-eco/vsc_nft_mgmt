package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

const (
	maxNameLength = 100  // maximum length if the name attribute (used by collections and nfts)
	maxDescLength = 1000 // maximum length if the description attribute (used by collections and nfts)
)

type Collection struct {
	ID           string      `json:"id"`
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

// --- Wasm exports (unchanged signatures) ---

//go:wasmexport col_create
func CreateCollection(payload *string) *string {
	return createCollectionImpl(payload, RealSDK{})
}

//go:wasmexport col_get
func GetCollection(id *string) *string {
	return getCollectionImpl(id, RealSDK{})
}

//go:wasmexport col_get_user
func GetCollectionsForOwner(owner *string) *string {
	return getCollectionsForOwnerImpl(owner, RealSDK{})
}

func createCollectionImpl(payload *string, chain SDKInterface) *string {
	// env := sdkInterface.GetEnv()
	input := FromJSON[CreateCollectionArgs](*payload, "collection args")

	input.Validate(chain)
	env := chain.GetEnv()
	creator := env.Sender.Address
	// if collectionExists(creator, input.Name) {
	// 	abortOnError(fmt.Errorf("collection with name '%s' already exists", input.Name), "")
	// }

	collectionId := newCollectionID(chain)

	collection := Collection{
		ID: strconv.FormatInt(collectionId, 10),

		Owner:        creator,
		Name:         input.Name,
		Description:  input.Description,
		CreationTxID: env.TxId,
	}
	saveCollection(&collection, chain)
	setCollectionCount(collectionId+int64(1), chain)
	return nil
}

func getCollectionImpl(id *string, chain SDKInterface) *string {
	collection := loadCollection(*id, chain)
	jsonStr := ToJSON(collection, "collection")
	return &jsonStr
}

func getCollectionsForOwnerImpl(owner *string, chain SDKInterface) *string {
	collectionIds := GetIDsFromIndex(CollectionsOwner+*owner, chain)
	collections := make([]Collection, 0)
	for _, n := range collectionIds {
		currentCollection := loadCollection(n, chain)
		collections = append(collections, *currentCollection)
	}
	jsonStr := ToJSON(collections, "collections")
	return &jsonStr
}

// Contract State Persistence
func saveCollection(collection *Collection, chain SDKInterface) error {
	b, err := json.Marshal(collection)
	if err != nil {
		chain.Abort("failed to marshal collection")
	}

	// save collection itself
	idKey := collectionKey(collection.ID)
	chain.StateSetObject(idKey, string(b))
	// save collection id into index for owner
	AddIDToIndex(CollectionsOwner+collection.Owner.String(), collection.ID, chain)

	return nil
}

func loadCollection(id string, chain SDKInterface) *Collection {
	if id == "" {
		chain.Abort("ID is mandatory")
	}
	key := collectionKey(id)
	ptr := chain.StateGetObject(key)
	if ptr == nil {
		chain.Abort(fmt.Sprintf("collection %s not found", id))
	}
	collection := FromJSON[Collection](*ptr, "collection")
	return collection
}

func (c *CreateCollectionArgs) Validate(chain SDKInterface) {
	if c.Name == "" {
		chain.Abort("name is mandatory")
	}
	if len(c.Name) > maxNameLength {
		chain.Abort(fmt.Sprintf("name: max %d chars", maxNameLength))
	}
	if len(c.Description) > maxDescLength {
		chain.Abort(fmt.Sprintf("desc: max %d chars", maxDescLength))
	}
}

func collectionKey(collectionId string) string {
	return fmt.Sprintf("col:%s", collectionId)
}

func newCollectionID(chain SDKInterface) int64 {
	return getCount(CollectionCount, chain)
}

func setCollectionCount(nextId int64, chain SDKInterface) {
	setCount(CollectionCount, nextId, chain)
}
