package main

import (
	"fmt"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

const (
	nftVersion         = 1   // for a possible versioning of the nft contract
	maxMetaKeys        = 25  // max count of metadata keys for an nft
	maxMetaKeyLength   = 50  // max length of a key within the metadata
	maxMetaValueLength = 512 // max length of a value within the metadata
)

// the basic nft object
type NFT struct {
	ID             uint64            `json:"id"`           // unique id of an nft
	Creator        sdk.Address       `json:"cr"`           // original creator of the nft
	Owner          sdk.Address       `json:"o"`            // current owner of the nft
	CreationTxID   string            `json:"txID"`         // tx when the nft was minted
	Collection     uint64            `json:"c"`            // current collection id the nft is part of
	Name           string            `json:"n"`            // general name of the nft
	Description    string            `json:"d"`            // long description of the nft
	Metadata       map[string]string `json:"m,omitempty"`  // additional metadata like properties, uri and more
	SingleTransfer bool              `json:"b"`            // true if the nft can only be transferred once
	EditionsTotal  uint32            `json:"et,omitempty"` // total number of editions (only set on genesis)
	Version        int               `json:"v"`            // version of the nft contract this nft was minted with
}

// EditionOverride stores ownership for a specific edition
type EditionOverride struct {
	Owner      sdk.Address `json:"owner"`
	Collection uint64      `json:"collection"`
}

type TransferNFTArgs struct {
	NftID      uint64      `json:"id"` // mandatory: id of the nft to get transferred
	Collection uint64      `json:"c"`  // mandatory: target collection
	Owner      sdk.Address `json:"o"`  // mandatory: target owner
}

type MintNFTArgs struct {
	Collection     *uint64           `json:"c"`     // mandatory: target collection id
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // opt: description
	SingleTransfer bool              `json:"bound"` // opt: non-transferrable (default false)
	Metadata       map[string]string `json:"meta"`  // opt: metadata
	EditionsTotal  uint32            `json:"et"`    // mandatory: number editions to mint
}

// MINT FUNCTIONS

// creation of NFT
//
//go:wasmexport nft_mint
func MintNFT(payload *string) *string {
	input := FromJSON[MintNFTArgs](*payload, "minting args")
	if input.Collection == nil {
		sdk.Abort("collection is mandatory")
	}

	collection := loadCollection(*input.Collection)
	creator := sdk.GetEnvKey("msg.sender")
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, *creator)

	// Ensure EditionsTotal is always at least 1
	et := input.EditionsTotal
	if et == 0 {
		et = 1
	}

	nftID := newNFTID()
	nft := &NFT{
		ID:             nftID,
		Creator:        sdk.Address(*creator),
		Owner:          sdk.Address(*creator),
		Version:        nftVersion,
		CreationTxID:   *sdk.GetEnvKey("tx.id"),
		Collection:     *input.Collection,
		Name:           input.Name,
		Description:    input.Description,
		Metadata:       input.Metadata,
		SingleTransfer: input.SingleTransfer,
		EditionsTotal:  et,
	}
	saveNFT(nft)

	EmitMintEvent(nftID, *creator, *creator, *input.Collection, et)

	// update NFT counter
	setCount(NFTsCount, nftID+uint64(1))
	return nil
}

// TRANSFER FUNCTIONS

// transfers an NFT between users or collections
//
//go:wasmexport nft_transfer
func TransferNFT(payload *string) *string {
	input := FromJSON[TransferNFTArgs](*payload, "transfer args")

	nft := loadNFT(input.NftID)
	editionIndex := uint32(0) // default for single NFTs
	owner, collection := nft.Owner, nft.Collection

	// Only multi-editions use override
	if nft.EditionsTotal > 1 {
		owner, collection = resolveEditionOwnerAndCollection(nft, editionIndex)
	}

	// prevent no-op transfers
	if owner == input.Owner && collection == input.Collection {
		sdk.Abort("source and target are the same")
	}

	collectionOnlyChange := owner == input.Owner && collection != input.Collection

	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()
	targetCollection := loadCollection(input.Collection)

	// validations
	// Owner transfer
	if owner != input.Owner {
		if *caller != *marketContract && *caller != owner.String() {
			sdk.Abort("only market or owner can transfer")
		}
		if nft.SingleTransfer && nft.Creator != owner {
			sdk.Abort("nft bound to owner")
		}
	}

	// Collection-only transfers
	if collectionOnlyChange {
		if *caller != owner.String() {
			sdk.Abort("only NFT owner can change collection")
		}
		if targetCollection.Owner != input.Owner {
			sdk.Abort("target collection not owned by caller")
		}
	}

	// Update ownership
	if nft.EditionsTotal > 1 {
		// For multi-editions, store override
		saveEditionOverride(nft.ID, editionIndex, input.Owner, input.Collection)

		EmitTransferEvent(
			UInt64ToString(nft.ID)+":"+strconv.FormatInt(int64(editionIndex), 10), // composite id nft:editionIndex
			owner.String(),
			input.Owner.String(),
			nft.Collection,
			input.Collection)
	} else {
		// Unique NFT transfer
		nft.Owner = input.Owner
		nft.Collection = input.Collection
		saveNFT(nft)

		EmitTransferEvent(
			UInt64ToString(nft.ID),
			owner.String(),
			nft.Owner.String(),
			collection,
			nft.Collection)
	}

	return nil
}

func saveEditionOverride(nftID uint64, editionIndex uint32, owner sdk.Address, collection uint64) {
	key := fmt.Sprintf("o:%d:%d", nftID, editionIndex)
	override := &EditionOverride{
		Owner:      owner,
		Collection: collection,
	}
	sdk.StateSetObject(key, ToJSON(override, "override"))
}

func loadEditionOverride(nftID uint64, editionIndex uint32) *EditionOverride {
	key := fmt.Sprintf("o:%d:%d", nftID, editionIndex)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		return nil
	}
	return FromJSON[EditionOverride](*ptr, "override")
}

func resolveEditionOwnerAndCollection(nft *NFT, editionIndex uint32) (sdk.Address, uint64) {
	override := loadEditionOverride(nft.ID, editionIndex)
	if override != nil {
		return override.Owner, override.Collection
	}
	return nft.Owner, nft.Collection
}

//go:wasmexport nft_burn
func BurnNFT(id *string) *string {
	nft := loadNFT(StringToUInt64(id))
	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()

	// validate burning permissions
	if *caller != *marketContract && *caller != nft.Owner.String() {
		sdk.Abort("only owner or market can burn")
	}

	// delete the NFT from state
	sdk.StateDeleteObject(nftKey(nft.ID))

	// emit burn event
	EmitBurnEvent(nft.ID, nft.Owner.String(), nft.Collection)
	return nil
}

// GET FUNCTIONS

// returns an NFT by ID, includes EditionsTotal
//
//go:wasmexport nft_get
func GetNFT(id *string) *string {
	nft := loadNFT(StringToUInt64(id))
	jsonStr := ToJSON(nft, "nft")
	return &jsonStr
}

// CONTRACT STATE INTERACTIONS

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

// VALIDATION & HELPERS

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
		sdk.Abort("name too long")
	}
	if len(description) > maxDescLength {
		sdk.Abort("description too long")
	}
	if collectionOwner.String() != caller {
		sdk.Abort("not the owner")
	}
	if len(metadata) > maxMetaKeys {
		sdk.Abort("too many meta keys")
	}
	for k, v := range metadata {
		if k == "" {
			sdk.Abort("meta keys empty")
		}
		if len(k) > maxMetaKeyLength {
			sdk.Abort("one meta key too long")
		}
		if len(v) > maxMetaValueLength {
			sdk.Abort("one meta value too long")
		}
	}
}

// // create and store NFT in state
// func createAndSaveNFT(
// 	nftId uint64,
// 	creator sdk.Address,
// 	owner sdk.Address,
// 	collection uint64,
// 	name string,
// 	description string,
// 	singleTransfer bool,
// 	metadata map[string]string,
// 	genesisEditionID *uint64,
// 	txId string,
// ) {
// 	var nftPrefs *NFTPrefs

// 	// first edition or unique NFT stores preferences
// 	if genesisEditionID == nil || *genesisEditionID == nftId {
// 		nftPrefs = &NFTPrefs{
// 			Name:           name,
// 			Description:    description,
// 			Metadata:       metadata,
// 			SingleTransfer: singleTransfer,
// 		}
// 	}

// 	nft := &NFT{
// 		ID:             nftId,
// 		Creator:        creator,
// 		Owner:          owner,
// 		Version:        nftVersion,
// 		CreationTxID:   txId,
// 		Collection:     collection,
// 		NFTPrefs:       nftPrefs,
// 		GenesisEdition: genesisEditionID,
// 	}
// 	saveNFT(nft)
// }

// generate state key for NFT
func nftKey(nftId uint64) string {
	return "n:" + strconv.FormatUint(nftId, 10)
}

// get next available NFT ID
func newNFTID() uint64 {
	return getCount(NFTsCount)
}
