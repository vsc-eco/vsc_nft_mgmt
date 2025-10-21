package main

import (
	"strconv"
	"strings"
	"vsc_nft_mgmt/sdk"
)

// Event represents a generic event emitted by the contract.
type Event struct {
	Type       string            `json:"t"`   // Type is the kind of event (e.g., "mint", "transfer").
	Attributes map[string]string `json:"att"` // Attributes are key/value pairs with event data.
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
func EmitTransferEvent(nftID string, fromAddress, toAddress string, fromCollection, toCollection string) {
	emitEvent("t", map[string]string{
		"id":   nftID,
		"f":    fromAddress,
		"t":    toAddress,
		"fCol": fromCollection,
		"tCol": toCollection,
	})
}

// EmitMintEvent emits an event for NFT minting.
// editionsTotal is included only if greater than zero.
func EmitMintEvent(nftID uint64, mintedByAddress, ownerCollection string, editionsTotal uint32) {
	attrs := map[string]string{
		"id":   UInt64ToString(nftID),
		"c":    mintedByAddress,
		"t":    strings.Split(ownerCollection, "/")[0],
		"tCol": strings.Split(ownerCollection, "/")[1],
	}
	if editionsTotal > 0 {
		attrs["edCnt"] = strconv.FormatInt(int64(editionsTotal), 10)
	}
	emitEvent("mint", attrs)
}

// EmitBurnEvent emits an event for burning an NFT or edition.
// The id may be a base NFT ID or an edition ID in the format "nftId:editionIndex".
func EmitBurnEvent(nftID string, ownerAddress string, collection string) {
	emitEvent("b", map[string]string{
		"id":  nftID,
		"c":   ownerAddress,
		"col": collection,
	})
}

// EmitCollectionCreatedEvent emits an event for creating a new collection.
func EmitCollectionCreatedEvent(collectionID uint64, createdByAddress string) {
	emitEvent("col", map[string]string{
		"id": UInt64ToString(collectionID),
		"c":  createdByAddress,
	})
}
