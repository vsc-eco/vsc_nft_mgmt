package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// Event represents a generic event emitted by the contract.
type Event struct {
	Type       string            `json:"type"`       // Type is the kind of event (e.g., "mint", "transfer").
	Attributes map[string]string `json:"attributes"` // Attributes are key/value pairs with event data.
}

// emitEvent constructs and logs an event as JSON.
func emitEvent(eventType string, attributes map[string]string) {
	event := Event{
		Type:       eventType,
		Attributes: attributes,
	}
	sdk.Log(ToJSON(event, eventType+" event data"))
}

// EmitTransferEvent emits an event for an NFT transfer.
// The id may be a base NFT ID or an edition ID in the format "nftId:editionIndex".
func EmitTransferEvent(nftID string, fromAddress, toAddress string, fromCollection, toCollection uint64) {
	emitEvent("transfer", map[string]string{
		"id":             nftID,
		"from":           fromAddress,
		"to":             toAddress,
		"fromCollection": UInt64ToString(fromCollection),
		"toCollection":   UInt64ToString(toCollection),
	})
}

// EmitMintEvent emits an event for NFT minting.
// editionsTotal is included only if greater than zero.
func EmitMintEvent(nftID uint64, mintedByAddress, receiverAddress string, collection uint64, editionsTotal uint32) {
	attrs := map[string]string{
		"id":           UInt64ToString(nftID),
		"by":           mintedByAddress,
		"to":           receiverAddress,
		"toCollection": UInt64ToString(collection),
	}
	if editionsTotal > 0 {
		attrs["editionsTotal"] = strconv.FormatInt(int64(editionsTotal), 10)
	}
	emitEvent("mint", attrs)
}

// EmitBurnEvent emits an event for burning an NFT or edition.
// The id may be a base NFT ID or an edition ID in the format "nftId:editionIndex".
func EmitBurnEvent(nftID string, ownerAddress string, collection uint64) {
	emitEvent("burn", map[string]string{
		"id":         nftID,
		"by":         ownerAddress,
		"collection": UInt64ToString(collection),
	})
}

// EmitCollectionCreatedEvent emits an event for creating a new collection.
func EmitCollectionCreatedEvent(collectionID uint64, createdByAddress string) {
	emitEvent("collection", map[string]string{
		"id": UInt64ToString(collectionID),
		"by": createdByAddress,
	})
}
