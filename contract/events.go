package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// Event is the common structure for all emitted events.
type Event struct {
	Type       string            `json:"type"`
	Attributes map[string]string `json:"attributes"`
}

// emitEvent builds the event and logs it as JSON.
func emitEvent(eventType string, attributes map[string]string) {
	event := Event{
		Type:       eventType,
		Attributes: attributes,
	}
	sdk.Log(ToJSON(event, eventType+" event data"))
}

// EmitTransferEvent emits a transfer event.
func EmitTransferEvent(nftID string, fromAddress, toAddress string, fromCollection, toCollection uint64) {
	emitEvent("transfer", map[string]string{
		"id":             nftID,
		"from":           fromAddress,
		"to":             toAddress,
		"fromCollection": UInt64ToString(fromCollection),
		"toCollection":   UInt64ToString(toCollection),
	})
}

// EmitMintEvent emits a mint event.
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

// EmitBurnEvent emits a burn event.
func EmitBurnEvent(nftID string, ownerAddress string, collection uint64) {
	emitEvent("burn", map[string]string{
		"id":         nftID,
		"by":         ownerAddress,
		"collection": UInt64ToString(collection),
	})
}

// EmitCollectionCreatedEvent emits a collection creation event.
func EmitCollectionCreatedEvent(collectionID uint64, createdByAddress string) {
	emitEvent("collection", map[string]string{
		"id": UInt64ToString(collectionID),
		"by": createdByAddress,
	})
}
