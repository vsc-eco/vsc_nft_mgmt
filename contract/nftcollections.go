package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"vsc_nft_mgmt/sdk"
)

const (
	maxNameLength        = 50
	maxDescriptionLength = 500
)

type NFTCollection struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"desc"`
	Owner        sdk.Address `json:"owner"`
	CreationTxID string      `json:"txid"`
}

// function arguments
type CreateNFTCollectionArgs struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}

//go:wasmexport col_create
func CreateNFTCollection(payload *string) *string {
	// env := sdkInterface.GetEnv()
	input, err := FromJSON[CreateNFTCollectionArgs](*payload)
	abortOnError(err, "invalid collection args")
	validationErrors := input.Validate()
	abortOnError(validationErrors, "validation failed")

	creator := getSenderAddress()
	// if collectionExists(creator, input.Name) {
	// 	abortOnError(fmt.Errorf("collection with name '%s' already exists", input.Name), "")
	// }

	collection := NFTCollection{
		ID:           generateUUID(),
		Owner:        creator,
		Name:         input.Name,
		Description:  input.Description,
		CreationTxID: getTxID(),
	}
	savingErrors := saveNFTCollection(&collection)
	abortOnError(savingErrors, "invalid collection args")

	// sdkInterface.Log(fmt.Sprintf("CreateNFTCollection: %s", collection.ID))
	return returnJsonResponse(
		true, map[string]interface{}{
			"id": collection.ID,
		},
	)
}

//go:wasmexport col_get
func GetCollection(id *string) *string {
	collection, err := loadNFTCollection(*id)
	abortOnError(err, "failed to load collection")
	jsonStr, err := ToJSON(collection)
	abortOnError(err, "failed to marshal collection")
	return &jsonStr
}

//go:wasmexport col_get_user
func GetNFTCollectionsForOwner(owner *string) *string {
	collectionIds := GetIDsFromIndex(CollectionsOwner + *owner)
	collections := make([]NFTCollection, 0)
	for _, n := range collectionIds {
		currentCollection, err := loadNFTCollection(n)
		abortOnError(err, "loading collection failed")
		collections = append(collections, *currentCollection)
	}
	jsonStr, err := ToJSON(collections)
	abortOnError(err, "failed to marshal collections")
	return &jsonStr
}

// Contract State Persistence
func saveNFTCollection(collection *NFTCollection) error {
	b, err := json.Marshal(collection)
	if err != nil {
		return err
	}

	// save collection itself
	idKey := collectionKey(collection.ID)
	getStore().Set(idKey, string(b))
	// save collection id into index for owner
	AddIDToIndex(CollectionsOwner+collection.Owner.String(), collection.ID)

	return nil
}

func loadNFTCollection(id string) (*NFTCollection, error) {
	if id == "" {
		return nil, fmt.Errorf("ID is mandatory")
	}
	key := collectionKey(id)
	ptr := getStore().Get(key)
	if ptr == nil {
		return nil, fmt.Errorf("collection %s not found", id)
	}
	collection, err := FromJSON[NFTCollection](*ptr)

	if err != nil {
		return nil, fmt.Errorf("failed unmarshal collection %s: %v", id, err)
	}
	return collection, nil
}

func (c *CreateNFTCollectionArgs) Validate() error {
	if c.Name == "" {
		return errors.New("name is mandatory")
	}
	if len(c.Name) > maxNameLength {
		return fmt.Errorf("name: max %d chars", maxNameLength)
	}
	if len(c.Description) > maxDescriptionLength {
		return fmt.Errorf("desc: max %d chars", maxDescriptionLength)
	}
	return nil
}
