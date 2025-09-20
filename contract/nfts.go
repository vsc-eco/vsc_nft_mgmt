package main

import (
	"fmt"
	"vsc_nft_mgmt/sdk"
)

const (
	nftVersion         = 1   // for a possible versioning of the nft contract
	maxMetaKeys        = 25  // maximum count of metadata keys for an nft
	maxMetaKeyLength   = 50  // maaximum length of a key within the metadata
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
	SingleTransfer bool        `json:"bound"` // true if the nft can only be transferred once (from creator to another user)
	NFTPrefs       *NFTPrefs   `json:"pref,omitempty"`
	Edition        *NFTEdition `json:"eds,omitempty"`

	// later more "NFT types" are possible like mutables and others
}

// non-unique nfts have additional edition data
type NFTEdition struct {
	EditionNumber  uint32 `json:"no"`
	EditionsTotal  uint32 `json:"t"`
	GenesisEdition uint64 `json:"g"` // only the genesis edition will have nftPrefs to save space
}

// more complex NFT preferences
// only get stored on unique nfts and genesis editions to reduce redundance within contract state
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
	Description    string            `json:"desc"`  // optional: description of the nft
	SingleTransfer bool              `json:"bound"` // optional: true if the nft is non-transferrable
	Metadata       map[string]string `json:"meta"`  // optional: additional metadata for the nft
}

type MintNFTEditionsArgs struct {
	Collection     uint64            `json:"c"`     // mandatory: collectionId
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // optional: description of the nft
	SingleTransfer bool              `json:"bound"` // optional: true if the nft is non-transferrable
	Metadata       map[string]string `json:"meta"`  // optional: additional metadata for the nft
	EditionsTotal  uint32            `json:"et"`    // mandatory: total count of editions
	GenesisEdition *uint64           `json:"g"`     // optional: set genesis nft id if editions should be extended
}

// MINT FUNCTIONS

// creation of an unique nft
//
//go:wasmexport nft_mint_unique
func MintNFTUnique(payload *string) *string {
	input := FromJSON[MintNFTArgs](*payload, "minting args")
	collection := loadCollection(input.Collection)
	creator := sdk.GetEnvKey("msg.sender")
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, *creator)
	nftID := newNFTID()
	env := sdk.GetEnv()

	createAndSaveNFT(
		nftID,
		sdk.Address(*creator),
		sdk.Address(*creator),
		input.Collection,
		input.Description,
		input.SingleTransfer,
		input.Metadata,
		false, 0, 0, false, // editionedNFT,editionNumber, genesisEditionID, extend editions
		env.TxId,
	)
	// add nft to index
	AddIDToIndex(NFTsCreator+*creator, nftID)
	AddIDToIndex(NFTsCollection+UInt64ToString(input.Collection), nftID)

	// increase count
	setCount(NFTsCount, nftID+1)
	return nil
}

// creation of an edition nft
// the genesis (first) nft of this series will holds potential "big" but redundant data
// all the following nfts have a property called "genesis" pointing to that genesis nft
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

	if input.GenesisEdition != nil {
		genesisEditionID = *input.GenesisEdition
		extendEditions = true
		existingEditionsCount = getEditionsCount(genesisEditionID)
	}

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

	// add genesis to creator index
	if !extendEditions {
		AddIDToIndex(NFTsCreator+*creator, genesisEditionID)
	}
	// batch-add to indices
	SliceToIndex(NFTsCollection+UInt64ToString(input.Collection), editionIDs)
	SliceToIndex(AllEditionsOfGenesis+UInt64ToString(genesisEditionID), editionIDs)
	SliceToIndex(AvailEditionsOfGenesis+UInt64ToString(genesisEditionID), editionIDs)
	// increase nft counter
	setCount(NFTsCount, nftID+uint64(len(editionIDs)))

	return nil
}

// TRANSFER FUNCTIONS

// transfers a nft between users and/or collections
// owner can move between collections
// market contract or owner can move to new owner
//
//go:wasmexport nft_transfer
func TransferNFT(payload *string) *string {

	input := FromJSON[TransferNFTArgs](*payload, "transfer args")
	nft := loadNFT(input.NftID)

	// avoid unnecessary non-transfer
	if nft.Owner == input.Owner && nft.Collection == input.Collection {
		sdk.Abort("source and target are the same")
	}

	caller := sdk.GetEnv().Caller // caller instead of sender to enable other contracts
	marketContract := getMarketContract()

	// if the nft should move from one user to another
	if input.Owner != nft.Owner {
		if caller != marketContract && caller != nft.Owner {
			sdk.Abort("only market or owner can transfer")
		}
		if nft.Creator != nft.Owner && nft.SingleTransfer {
			// nft moved once after mint and is non-transferrable
			sdk.Abort("nft bound to owner")
		}
	} else {
		// it is a move between collections of the owner
		if caller != nft.Owner {
			sdk.Abort("only owner can move")
		}
	}
	// check if the target owner owns the target collection
	collection := loadCollection(input.Collection)
	if collection.Owner != input.Owner {
		sdk.Abort("collection not owned by new owner")
	}

	originalCollection := nft.Collection
	nft.Collection = input.Collection
	nft.Owner = input.Owner
	saveNFT(nft)

	// remove the nft from the source collection index
	RemoveIDFromIndex(NFTsCollection+UInt64ToString(originalCollection), nft.ID)
	// add the nft to the target collection index
	AddIDToIndex(NFTsCollection+UInt64ToString(nft.Collection), nft.ID)
	// if it is an edition: remove the nft from the list of available editions (if still in there)
	if nft.Edition != nil {
		RemoveIDFromIndex(AvailEditionsOfGenesis+UInt64ToString(nft.Edition.GenesisEdition), nft.ID)
	}
	sdk.Log("nft transferred to new owner / moved to other collection")
	return nil

}

// GET FUNCTIONS
// returns an nft for a given nft id
//
//go:wasmexport nft_get
func GetNFT(id *string) *string {
	// get one NFT by Id
	nft := loadNFT(StringToUInt64(id))
	jsonStr := ToJSON(nft, "nft")
	sdk.Log(jsonStr)
	return &jsonStr
}

// returns the next available edition nft still in creators collection for a given genesis nft id
//
//go:wasmexport nft_get_available
func GetNextAvailableEditionForNFT(id *string) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AvailEditionsOfGenesis + *id)
	// get the smalles nftId
	min := nftIds[0]
	for _, v := range nftIds[1:] {
		if v < min {
			min = v
		}
	}
	minStr := UInt64ToString(min)
	return &minStr
}

// Contract State Interactions

// stores an nft (minded / updated)
func saveNFT(nft *NFT) {
	key := nftKey(nft.ID)
	b := ToJSON(nft, "nft")
	sdk.StateSetObject(key, string(b))

}

// returns a single nft by id
func loadNFT(id uint64) *NFT {
	key := nftKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("nft %d not found", id))
	}
	nft := FromJSON[NFT](*ptr, "nft")
	if nft.Edition != nil {
		nft.Edition.EditionsTotal = getEditionsCount(id)
	}
	return nft
}

func getEditionsCount(id uint64) uint32 {
	editionsTotal := GetIDsFromIndex(AllEditionsOfGenesis + UInt64ToString(id))
	return uint32(len(editionsTotal))
}

// functions arguments validation
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
		return
	}
	if len(description) > maxDescLength {
		sdk.Abort(fmt.Sprintf("desc max. %d chars", maxDescLength))
	}
	if collectionOwner.String() != caller {
		sdk.Abort("not the owner")
	}

	// check size of the metadata to avoid bloat of the state storage
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

	// if we mint an nft edition
	if editionedNFT {
		nftEdition = &NFTEdition{
			EditionNumber:  editionNumber,
			GenesisEdition: genesisEditionID,
		}
	}
	// if it is the first edition, extended editions or a unique nft
	if !editionedNFT || !extendEditions || editionNumber == 1 {
		// Only genesis edition keeps prefs to avoid redundant state storage
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

func nftKey(nftId uint64) string {
	return fmt.Sprintf("n:%d", nftId)
}

func newNFTID() uint64 {
	return getCount(NFTsCount)
}
