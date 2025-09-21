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
	GenesisEdition *uint64     `json:"g"` // only the genesis edition will have nftPrefs
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

	emitMintEvent(nftID, *creator, *creator, input.Collection, nil)

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

	// check if we are extending an existing genesis
	if input.GenesisEdition != nil {
		genesisEditionID = *input.GenesisEdition
		extendEditions = true
	}

	// create all editions
	for i := 0; i < int(input.EditionsTotal); i++ {
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
		emitMintEvent(nftID+uint64(i), *creator, *creator, input.Collection, &genesisEditionID)
	}

	// update NFT counter
	setCount(NFTsCount, nftID+uint64(input.EditionsTotal))
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
	nft := loadNFT(input.NftID)
	original := nft

	// prevent no-op transfers
	if original.Owner == input.Owner && original.Collection == input.Collection {
		sdk.Abort("source and target are the same")
	}

	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()

	// validate transfer permissions
	if input.Owner != original.Owner {
		if caller != marketContract && *caller != original.Owner.String() {
			sdk.Abort("only market or owner can transfer")
		}
		if original.Creator != original.Owner && original.SingleTransfer {
			sdk.Abort("nft bound to owner")
		}
	} else {
		if *caller != original.Owner.String() {
			sdk.Abort("only owner can move")
		}
	}

	collection := loadCollection(input.Collection)
	if collection.Owner != input.Owner {
		sdk.Abort("collection not owned by new owner")
	}

	nft.Collection = input.Collection
	nft.Owner = input.Owner
	saveNFT(nft)

	emitTransferEvent(nft.ID, original.Owner.String(), nft.Owner.String(), original.Collection, nft.Collection)

	return nil
}

// --------------------
// GET FUNCTIONS
// --------------------

// returns an NFT by ID, includes EditionsTotal
//
//go:wasmexport nft_get
func GetNFT(id *string) *string {
	nft := loadNFT(StringToUInt64(id))
	jsonStr := ToJSON(nft, "nft")
	return &jsonStr
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

// NFT loader
func loadNFT(id uint64) *NFT {
	key := nftKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("nft %d not found", id))
	}
	nft := FromJSON[NFT](*ptr, "nft")
	return nft
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
	var nftPrefs *NFTPrefs

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
		GenesisEdition: &genesisEditionID,
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

func emitTransferEvent(nftId uint64, fromAddress string, toAddress string, fromCollection uint64, toCollection uint64) {
	sdk.Log(fmt.Sprintf(
		"Transfer|tokenId:%d|from:%s|to:%s|fromCollection:%d|toCollection:%d",
		nftId,
		fromAddress,
		toAddress,
		fromCollection,
		toCollection,
	))
}

func emitMintEvent(nftId uint64, mindedByAddress string, receiverAddress string, collection uint64, genesisNFTID *uint64) {
	event := fmt.Sprintf(
		"Mint|tokenId:%d|by:%s|to:%s|collection:%d",
		nftId,
		mindedByAddress,
		receiverAddress,
		collection,
	)

	if genesisNFTID != nil {
		event += fmt.Sprintf("|genesis:%d", *genesisNFTID)
	}

	sdk.Log(event)
}
