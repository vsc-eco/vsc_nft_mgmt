package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"vsc_nft_mgmt/sdk"
)

const (
	maxNFTNameLength        = 200
	maxNFTDescriptionLength = 1000
	nftVersion              = 1
	maxMetadataKeys         = 25
	maxMetadataKeyLength    = 50
	maxMetadataValueLength  = 512
)

// the basic nft object
type NFT struct {
	ID           string      `json:"id"`
	Creator      sdk.Address `json:"creator"`
	Owner        sdk.Address `json:"owner"`
	Version      int         `json:"version"`
	CreationTxID string      `json:"txID"`
	Collection   string      `json:"col"`
	Transferable bool        `json:"transfer"` // only one transfer possible (from creator to another user)
	NFTPrefs     *NFTPrefs   `json:"prefs,omitempty"`
	Edition      *NFTEdition `json:"edition,omitempty"`

	// later more "NFT types" are possible like mutables and others
}

// non-unique nfts have additional edition data
type NFTEdition struct {
	EditionNumber  int64  `json:"no"`
	EditionsTotal  int64  `json:"total"`
	GenesisEdition string `json:"genesis"` // only the genesis edition will have nftPrefs to save space
}

// more complex NFT preferences
// only get stored on unique nfts and genesis editions to reduce redundance within contract state
type NFTPrefs struct {
	Description string            `json:"desc"`
	Metadata    map[string]string `json:"meta,omitempty"`
}

type TransferNFTArgs struct {
	NftID      string      `json:"id"`
	Collection string      `json:"col"`
	Owner      sdk.Address `json:"owner"`
}

type MintNFTArgs struct {
	Collection   string            `json:"col"`
	Name         string            `json:"name"`
	Description  string            `json:"desc"`
	Transferable bool              `json:"transfer"`
	Metadata     map[string]string `json:"meta"`
}

type MintNFTEditionsArgs struct {
	Collection    string            `json:"col"`
	Name          string            `json:"name"`
	Transferable  bool              `json:"transfer"`
	EditionsTotal int64             `json:"editionsTotal"`
	Metadata      map[string]string `json:"meta"`
	Description   string            `json:"desc"`
}

// exported TRANSFER function
//
//go:wasmexport nft_transfer
func TransferNFT(payload *string) *string {
	// owner can move between collections
	// market contract or owner can move to new owner

	input, err := FromJSON[TransferNFTArgs](*payload)
	abortOnError(err, "invalid args")

	nft, err := loadNFT(input.NftID)
	abortOnError(err, "loading nft failed")

	// avoid unnecessary non-transfer
	if nft.Owner == input.Owner && nft.Collection == input.Collection {
		abortOnError(err, "source and target are the same")
	}

	caller := getCallerAddress() // caller instead of sender to enable other contracts
	marketContract, err := getMarketContract()
	abortOnError(err, "loading market failed")

	_, errCollection := loadNFTCollection(input.Collection)
	abortOnError(errCollection, "loading collection failed")

	// if the nft should move from one user to another
	if input.Owner != nft.Owner {
		if caller != marketContract && caller != nft.Owner {
			abortCustom("only market or owner can transfer")
		}
		if nft.Creator != nft.Owner && !nft.Transferable {
			// nft moved once after mint and is non-transferrable
			abortCustom("nft bound to owner")
		}
	} else {
		// it is a move between collections of the owner
		if caller != nft.Owner {
			abortCustom("only owner can move")
		}
	}

	originalCollection := nft.Collection
	nft.Collection = input.Collection
	nft.Owner = input.Owner
	saveNFT(nft)

	// remove the nft from the source collection index
	RemoveIDFromIndex(NFTsCollection+originalCollection, nft.ID)
	// add the nft to the target collection index
	AddIDToIndex(NFTsCollection+nft.Collection, nft.ID)
	// if it is an edition: remove the nft from the list of available editions (if still in there)
	if nft.Edition != nil {
		RemoveIDFromIndex(AvailEditionsOfGenesis+nft.Edition.GenesisEdition, nft.ID)
	}
	return returnJsonResponse(
		true, map[string]interface{}{
			"id": input.NftID,
		},
	)

}

// exported MINT functions

//go:wasmexport nft_mint_unique
func MintNFTUnique(payload *string) *string {
	input, err := FromJSON[MintNFTArgs](*payload)
	abortOnError(err, "invalid minting args")

	collection, err := loadNFTCollection(input.Collection)
	abortOnError(err, "loading collection failed")

	creator := getSenderAddress()
	abortOnError(validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, creator), "validation failed")

	nft, err := createAndSaveNFT(
		creator,
		creator,
		input.Collection,
		input.Description,
		input.Transferable,
		input.Metadata,
		0, 0, "", // editionNumber, editionsTotal, genesisEditionID
	)
	abortOnError(err, "creating NFT failed")

	return returnJsonResponse(
		true, map[string]interface{}{
			"id": nft.ID,
		},
	)
}

//go:wasmexport nft_mint_edition
func MintNFTEditions(payload *string) *string {
	input, err := FromJSON[MintNFTEditionsArgs](*payload)
	abortOnError(err, "invalid minting args")

	collection, err := loadNFTCollection(input.Collection)
	abortOnError(err, "loading collection failed")

	creator := getSenderAddress()
	abortOnError(validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, creator), "validation failed")

	if input.EditionsTotal <= 0 {
		abortOnError(errors.New("editions not set"), "invalid editions total")
	}

	var genesisEditionID string
	for editionNumber := 1; editionNumber <= int(input.EditionsTotal); editionNumber++ {
		nft, err := createAndSaveNFT(
			creator,
			creator,
			input.Collection,
			input.Description,
			input.Transferable,
			input.Metadata,
			int64(editionNumber),
			input.EditionsTotal,
			genesisEditionID,
		)
		abortOnError(err, fmt.Sprintf("creating edition %d failed", editionNumber))

		if editionNumber == 1 {
			genesisEditionID = nft.ID
		}
	}

	return returnJsonResponse(
		true, map[string]interface{}{
			"id": genesisEditionID,
		})
}

// exported GET functions

//go:wasmexport nft_get
func GetNFT(id *string) *string {
	// get one NFT by Id
	nft, err := loadNFT(*id)
	abortOnError(err, "failed to load nft")
	jsonStr, err := ToJSON(nft)
	abortOnError(err, "failed to marshal nft")
	return &jsonStr
}

//go:wasmexport nft_get_collection
func GetNFTsForCollection(collectionId *string) *string {
	// get all NFTs in a collection
	nftIds := GetIDsFromIndex(NFTsCollection + *collectionId)
	jsonStr := getNFTsByIds(nftIds, false)
	return &jsonStr
}

//go:wasmexport nft_get_owner
func GetNFTsForOwner(owner *string) *string {
	// get all NFTs owned by a user
	collectionIds := GetIDsFromIndex(CollectionsOwner + *owner)
	nfts := make([]*NFT, 0, len(collectionIds))
	// iterate over all collections of the user
	for _, n := range collectionIds {
		nftIds := GetIDsFromIndex(NFTsCollection + n)

		// iterate over all nfts in the collection
		for _, n := range nftIds {
			currentNFT, err := loadNFT(n)
			abortOnError(err, "loading nft failed")
			nfts = append(nfts, currentNFT)
		}
	}
	jsonStr, err := ToJSON(nfts)
	abortOnError(err, "failed to marshal nfts")
	return &jsonStr
}

//go:wasmexport nft_get_creator
func GetNFTsForCreator(creator *string) *string {
	// get all NFTs created by a user
	nftIds := GetIDsFromIndex(NFTsCreator + *creator)
	jsonStr := getNFTsByIds(nftIds, true)
	return &jsonStr
}

//go:wasmexport nft_get_editions
func GetEditionsForNFT(id *string) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AllEditionsOfGenesis + *id)
	jsonStr := getNFTsByIds(nftIds, false)
	return &jsonStr
}

//go:wasmexport nft_get_available
func GetAvailableEditionsForNFT(id *string) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AvailEditionsOfGenesis + *id)
	jsonStr := getNFTsByIds(nftIds, false)
	return &jsonStr
}

func getNFTsByIds(nftIds []string, withRelatedEditions bool) string {
	nfts := make([]*NFT, 0, len(nftIds))
	// iterate over all nfts
	for _, n := range nftIds {
		currentNFT, err := loadNFT(n)
		abortOnError(err, "loading nft failed")
		nfts = append(nfts, currentNFT)
		if withRelatedEditions {
			// iterate over all editions of genesis edition
			editionIds := GetIDsFromIndex(AllEditionsOfGenesis + n)
			for _, e := range editionIds {
				currentEdition, err := loadNFT(e)
				abortOnError(err, "loading edition failed")
				nfts = append(nfts, currentEdition)
			}
		}
	}

	jsonStr, err := ToJSON(nfts)
	abortOnError(err, "failed to marshal nfts")
	return jsonStr
}

// Contract State Persistence

// stores an nft (minded / updated)
func saveNFT(nft *NFT) error {
	key := nftKey(nft.ID)
	b, err := json.Marshal(nft)
	if err != nil {
		return err
	}
	getStore().Set(key, string(b))
	return nil
}

// returns a single nft by id
func loadNFT(id string) (*NFT, error) {
	key := nftKey(id)
	ptr := getStore().Get(key)
	if ptr == nil {
		return nil, fmt.Errorf("nft %s not found", id)
	}
	nft, err := FromJSON[NFT](*ptr)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal nft %s: %v", id, err)
	}
	return nft, nil
}

// functions arguments validation
func validateMintArgs(
	name string,
	description string,
	metadata map[string]string,
	collectionOwner sdk.Address,
	caller sdk.Address,
) error {
	if name == "" {
		return errors.New("name is mandatory")
	}
	if len(name) > maxNFTNameLength {
		return fmt.Errorf("name can only be %d characters long", maxNFTNameLength)
	}
	if len(description) > maxNFTDescriptionLength {
		return fmt.Errorf("description can only be %d characters long", maxNFTDescriptionLength)
	}
	if collectionOwner != caller {
		return errors.New("collection owner does not match")
	}

	// check size of the metadata to avoid bloat of the state storage
	if len(metadata) > maxMetadataKeys {
		return fmt.Errorf("metadata can contain at most %d keys", maxMetadataKeys)
	}
	for k, v := range metadata {
		if k == "" {
			return errors.New("metadata keys cannot be empty")
		}
		if len(k) > maxMetadataKeyLength {
			return fmt.Errorf("metadata key '%s' exceeds max length of %d", k, maxMetadataKeyLength)
		}
		if len(v) > maxMetadataValueLength {
			return fmt.Errorf("metadata value for key '%s' exceeds max length of %d", k, maxMetadataValueLength)
		}
	}

	return nil
}

func createAndSaveNFT(
	creator sdk.Address,
	owner sdk.Address,
	collection string,
	description string,
	transferable bool,
	metadata map[string]string,
	editionNumber int64,
	editionsTotal int64,
	genesisEditionID string,
) (*NFT, error) {
	nftID := generateUUID()

	var nftEdition *NFTEdition
	var nftPrefs *NFTPrefs

	if editionNumber > 0 {
		nftEdition = &NFTEdition{
			EditionNumber:  editionNumber,
			EditionsTotal:  editionsTotal,
			GenesisEdition: genesisEditionID,
		}
		if editionNumber == 1 {
			// Only genesis edition keeps prefs
			nftPrefs = &NFTPrefs{
				Description: description,
				Metadata:    metadata,
			}
		}
	} else {
		// Unique NFT or genesis
		nftPrefs = &NFTPrefs{
			Description: description,
			Metadata:    metadata,
		}
	}

	nft := &NFT{
		ID:           nftID,
		Creator:      creator,
		Owner:        owner,
		Version:      nftVersion,
		CreationTxID: getTxID(),
		Collection:   collection,
		Transferable: transferable,
		NFTPrefs:     nftPrefs,
		Edition:      nftEdition,
	}
	saveNFT(nft)
	if nft.NFTPrefs != nil {
		// only store unique nfts or genesis editions to creator index
		AddIDToIndex(NFTsCreator+nft.Creator.String(), nftID)
	}
	if nft.Edition != nil {
		// add editions to genesis nft index
		AddIDToIndex(AllEditionsOfGenesis+nft.Edition.GenesisEdition, nftID)
		AddIDToIndex(AvailEditionsOfGenesis+nft.Edition.GenesisEdition, nftID)

	}
	return nft, nil
}
