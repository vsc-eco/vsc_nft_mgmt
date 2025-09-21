package main

import (
	"fmt"
	"vsc_nft_mgmt/sdk"
)

const (
	nftVersion         = 1   // for a possible versioning of the nft contract
	maxMetaKeys        = 25  // maximum count of metadata keys for an nft
	maxMetaKeyLength   = 50  // maximum length of a key within the metadata
	maxMetaValueLength = 512 // maximum length of a value within the metadata
	maxEditions        = 100 // maximum editions mintable at once
)

// the basic nft object
type NFT struct {
	ID             uint64      `json:"id"`
	Creator        sdk.Address `json:"creator"`
	Owner          sdk.Address `json:"owner"`
	Version        int         `json:"v"`
	CreationTxID   string      `json:"txID"`
	Collection     uint64      `json:"c"`
	SingleTransfer bool        `json:"bound"` // true if the nft can only be transferred once
	NFTPrefs       *NFTPrefs   `json:"pref,omitempty"`
	Edition        *NFTEdition `json:"eds,omitempty"` // only present for editions
}

// non-unique nfts have additional edition data
type NFTEdition struct {
	EditionNumber  uint32 `json:"no"`
	EditionsTotal  uint32 `json:"t"`
	GenesisEdition uint64 `json:"g"` // only the genesis edition will have nftPrefs
}

// more complex NFT preferences
// only stored on unique NFTs and genesis editions to reduce redundant state storage
type NFTPrefs struct {
	Description string            `json:"desc"`
	Metadata    map[string]string `json:"meta,omitempty"`
}

type TransferNFTArgs struct {
	NftID      uint64      `json:"id"`
	Collection uint64      `json:"c"`
	Owner      sdk.Address `json:"owner"`
}

type MintNFTArgs struct {
	Collection     uint64            `json:"c"`     // mandatory: collectionId
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // optional: description
	SingleTransfer bool              `json:"bound"` // optional: true if non-transferrable
	Metadata       map[string]string `json:"meta"`  // optional: additional metadata
}

type MintNFTEditionsArgs struct {
	Collection     uint64            `json:"c"`     // mandatory: collectionId
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // optional: description
	SingleTransfer bool              `json:"bound"` // optional: non-transferrable
	Metadata       map[string]string `json:"meta"`  // optional: metadata
	EditionsTotal  uint32            `json:"et"`    // mandatory: total editions to mint
	GenesisEdition *uint64           `json:"g"`     // optional: existing genesis id to extend editions
}

// --------------------
// MINT FUNCTIONS
// --------------------

// creation of a unique NFT
//
//go:wasmexport nft_mint_unique
func MintNFTUnique(payload *string) *string {
	input := FromJSON[MintNFTArgs](*payload, "minting args")
	collection := loadCollection(input.Collection)
	creator := sdk.GetEnvKey("msg.sender")

	// validate input fields
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, *creator)
	nftID := newNFTID()
	env := sdk.GetEnv()

	// create and save NFT
	createAndSaveNFT(
		nftID,
		sdk.Address(*creator),
		sdk.Address(*creator),
		input.Collection,
		input.Description,
		input.SingleTransfer,
		input.Metadata,
		false, 0, 0, false, // editionedNFT, editionNumber, genesisEditionID, extendEditions
		env.TxId,
	)

	// add NFT to indexes: creator and collection
	AddIDToIndex(NFTsCreator+*creator, nftID)
	AddIDToIndex(NFTsCollection+UInt64ToString(input.Collection), nftID)

	// increment global NFT counter
	setCount(NFTsCount, nftID+1)
	return nil
}

// creation of NFT editions
//
//go:wasmexport nft_mint_edition
func MintNFTEditions(payload *string) *string {
	input := FromJSON[MintNFTEditionsArgs](*payload, "minting args")
	if input.EditionsTotal > maxEditions {
		sdk.Abort(fmt.Sprintf("%d can be minted at once", maxEditions))
	}
	collection := loadCollection(input.Collection)
	creator := sdk.GetEnvKey("msg.sender")
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, *creator)

	if input.EditionsTotal <= 0 {
		sdk.Abort("editions total <= 0")
	}

	txId := sdk.GetEnvKey("tx.id")
	nftID := newNFTID()
	genesisEditionID := nftID

	existingEditionsCount := uint32(0)
	extendEditions := false
	var editionIDs []uint64

	// check if we are extending an existing genesis
	if input.GenesisEdition != nil {
		genesisEditionID = *input.GenesisEdition
		extendEditions = true
		// needed to compute offsets for edition numbering
		existingEditionsCount = getEditionsCount(genesisEditionID)
	}

	// create all editions
	for i := 1; i <= int(input.EditionsTotal); i++ {
		createAndSaveNFT(
			nftID+uint64(i),
			sdk.Address(*creator),
			sdk.Address(*creator),
			input.Collection,
			input.Description,
			input.SingleTransfer,
			input.Metadata,
			true,
			existingEditionsCount+uint32(i),
			genesisEditionID,
			extendEditions,
			*txId,
		)
		editionIDs = append(editionIDs, nftID+uint64(i))
	}

	// add genesis to creator index if new
	if !extendEditions {
		AddIDToIndex(NFTsCreator+*creator, genesisEditionID)
	}

	// batch-add editions to collection and edition indices (efficient)
	SliceToIndex(NFTsCollection+UInt64ToString(input.Collection), editionIDs)
	SliceToIndex(AllEditionsOfGenesis+UInt64ToString(genesisEditionID), editionIDs)
	SliceToIndex(AvailEditionsOfGenesis+UInt64ToString(genesisEditionID), editionIDs)

	// update NFT counter
	setCount(NFTsCount, nftID+uint64(len(editionIDs)))
	return nil
}

// --------------------
// TRANSFER FUNCTIONS
// --------------------

// transfers an NFT between users or collections
//
//go:wasmexport nft_transfer
func TransferNFT(payload *string) *string {
	input := FromJSON[TransferNFTArgs](*payload, "transfer args")

	// lightweight loader that skips EditionsTotal computation for gas savings
	nft := loadNFTNoCount(input.NftID)

	// prevent no-op transfers
	if nft.Owner == input.Owner && nft.Collection == input.Collection {
		sdk.Abort("source and target are the same")
	}

	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()

	// validate transfer permissions
	if input.Owner != nft.Owner {
		if caller != marketContract && *caller != nft.Owner.String() {
			sdk.Abort("only market or owner can transfer")
		}
		if nft.Creator != nft.Owner && nft.SingleTransfer {
			sdk.Abort("nft bound to owner")
		}
	} else {
		if *caller != nft.Owner.String() {
			sdk.Abort("only owner can move")
		}
	}

	collection := loadCollection(input.Collection)
	if collection.Owner != input.Owner {
		sdk.Abort("collection not owned by new owner")
	}

	originalCollection := nft.Collection
	nft.Collection = input.Collection
	nft.Owner = input.Owner
	saveNFT(nft)

	RemoveIDFromIndex(NFTsCollection+UInt64ToString(originalCollection), nft.ID)
	AddIDToIndex(NFTsCollection+UInt64ToString(nft.Collection), nft.ID)

	// remove from available editions if applicable
	if nft.Edition != nil {
		RemoveIDFromIndex(AvailEditionsOfGenesis+UInt64ToString(nft.Edition.GenesisEdition), nft.ID)
	}

	return nil
}

// --------------------
// GET FUNCTIONS
// --------------------

// returns an NFT by ID, includes EditionsTotal
//
//go:wasmexport nft_get
func GetNFT(id *string) *string {
	nft := loadNFTWithCount(StringToUInt64(id))
	jsonStr := ToJSON(nft, "nft")
	sdk.Log(jsonStr)
	return &jsonStr
}

// returns the next available edition for a given genesis NFT
//
//go:wasmexport nft_get_available
func GetNextAvailableEditionForNFT(id *string) *string {
	nftIds := GetIDsFromIndex(AvailEditionsOfGenesis + *id)
	if len(nftIds) == 0 {
		return nil
	}
	min := nftIds[0]
	for _, v := range nftIds[1:] {
		if v < min {
			min = v
		}
	}
	minStr := UInt64ToString(min)
	return &minStr
}

// --------------------
// CONTRACT STATE INTERACTIONS
// --------------------

// store an NFT in state
func saveNFT(nft *NFT) {
	key := nftKey(nft.ID)
	b := ToJSON(nft, "nft")
	sdk.StateSetObject(key, string(b))
}

// load NFT and compute EditionsTotal
func loadNFTWithCount(id uint64) *NFT {
	nft := loadNFTBase(id)
	if nft.Edition != nil {
		nft.Edition.EditionsTotal = getEditionsCount(nft.Edition.GenesisEdition)
	}
	return nft
}

// load NFT without computing EditionsTotal (cheaper)
func loadNFTNoCount(id uint64) *NFT {
	return loadNFTBase(id)
}

// core NFT loader
func loadNFTBase(id uint64) *NFT {
	key := nftKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("nft %d not found", id))
	}
	nft := FromJSON[NFT](*ptr, "nft")
	return nft
}

// compute editions count from AllEditionsOfGenesis index
func getEditionsCount(genesisID uint64) uint32 {
	editionsTotal := GetIDsFromIndex(AllEditionsOfGenesis + UInt64ToString(genesisID))
	return uint32(len(editionsTotal))
}

// --------------------
// VALIDATION & HELPERS
// --------------------

func validateMintArgs(
	name string,
	description string,
	metadata map[string]string,
	collectionOwner sdk.Address,
	caller string,
) {
	if name == "" {
		sdk.Abort("name is mandatory")
	}
	if len(name) > maxNameLength {
		sdk.Abort(fmt.Sprintf("name must between 1 - %d chars", maxNameLength))
	}
	if len(description) > maxDescLength {
		sdk.Abort(fmt.Sprintf("desc max. %d chars", maxDescLength))
	}
	if collectionOwner.String() != caller {
		sdk.Abort("not the owner")
	}
	if len(metadata) > maxMetaKeys {
		sdk.Abort(fmt.Sprintf("meta max. %d keys", maxMetaKeys))
	}
	for k, v := range metadata {
		if k == "" {
			sdk.Abort("meta keys empty")
		}
		if len(k) > maxMetaKeyLength {
			sdk.Abort(fmt.Sprintf("meta key '%s' > %d chars", k, maxMetaKeyLength))
		}
		if len(v) > maxMetaValueLength {
			sdk.Abort(fmt.Sprintf("meta value for '%s' > %d chars", k, maxMetaValueLength))
		}
	}
}

// create and store NFT in state
func createAndSaveNFT(
	nftId uint64,
	creator sdk.Address,
	owner sdk.Address,
	collection uint64,
	description string,
	singleTransfer bool,
	metadata map[string]string,
	editionedNFT bool,
	editionNumber uint32,
	genesisEditionID uint64,
	extendEditions bool,
	txId string,
) {
	var nftEdition *NFTEdition
	var nftPrefs *NFTPrefs

	// set edition info if applicable
	if editionedNFT {
		nftEdition = &NFTEdition{
			EditionNumber:  editionNumber,
			GenesisEdition: genesisEditionID,
		}
	}

	// first edition or unique NFT stores preferences
	if !editionedNFT || !extendEditions || editionNumber == 1 {
		nftPrefs = &NFTPrefs{
			Description: description,
			Metadata:    metadata,
		}
	}

	nft := &NFT{
		ID:             nftId,
		Creator:        creator,
		Owner:          owner,
		Version:        nftVersion,
		CreationTxID:   txId,
		Collection:     collection,
		SingleTransfer: singleTransfer,
		NFTPrefs:       nftPrefs,
		Edition:        nftEdition,
	}
	saveNFT(nft)
}

// generate state key for NFT
func nftKey(nftId uint64) string {
	return fmt.Sprintf("n:%d", nftId)
}

// get next available NFT ID
func newNFTID() uint64 {
	return getCount(NFTsCount)
}
