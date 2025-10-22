package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// emitEventJSON is the centralized, gas-optimized event emitter.
// `attributesJSON` must be a valid JSON object string (e.g. {"id":1,"cr":"alice"}).
func emitEventJSON(eventType string, attributesJSON string) {
	txID := sdk.GetEnvKey("tx.id")
	b := make([]byte, 0, len(eventType)+len(attributesJSON)+len(*txID)+32)

	// {"type":"mint","attributes":{...},"tx":"123"}
	b = append(b, '{')

	// "type":"..."
	b = append(b, '"', 't', 'y', 'p', 'e', '"', ':', '"')
	b = append(b, eventType...)
	b = append(b, '"', ',')

	// "attributes":{...}
	b = append(b, '"', 'a', 't', 't', 'r', 'i', 'b', 'u', 't', 'e', 's', '"', ':')
	b = append(b, attributesJSON...)
	b = append(b, ',')

	// "tx":"123"
	b = append(b, '"', 't', 'x', '"', ':', '"')
	b = append(b, (*txID)...)
	b = append(b, '"', '}')

	s := string(b)
	sdk.Log(s)
}

// EmitMintEvent: {"id":123,"cr":"addr","oc":"owner_col","ed":100}
func EmitMintEvent(id uint64, creator string, ownerCol string, editions uint32) {
	attrs := make([]byte, 0, len(creator)+len(ownerCol)+32)
	attrs = append(attrs, '{')

	// "id":123
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// ,"cr":"..."
	attrs = append(attrs, ',', '"', 'c', 'r', '"', ':', '"')
	attrs = append(attrs, creator...)
	attrs = append(attrs, '"')

	// ,"oc":"..."
	attrs = append(attrs, ',', '"', 'o', 'c', '"', ':', '"')
	attrs = append(attrs, ownerCol...)
	attrs = append(attrs, '"')

	// ,"ed":100
	attrs = append(attrs, ',', '"', 'e', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, uint64(editions), 10)

	attrs = append(attrs, '}')

	emitEventJSON("mint", string(attrs))
}

// emitTransfer: {"id":123,"ed":1,"fr":"fromAddr","to":"toAddr"}
func emitTransfer(id uint64, ed *uint32, from string, to string) {
	attrs := make([]byte, 0, len(from)+len(to)+48)
	attrs = append(attrs, '{')

	// "id":123
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// optional edition
	if ed != nil {
		attrs = append(attrs, ',', '"', 'e', 'd', '"', ':')
		attrs = strconv.AppendUint(attrs, uint64(*ed), 10)
	}

	// ,"fr":"..."
	attrs = append(attrs, ',', '"', 'f', 'r', '"', ':', '"')
	attrs = append(attrs, from...)
	attrs = append(attrs, '"')

	// ,"to":"..."
	attrs = append(attrs, ',', '"', 't', 'o', '"', ':', '"')
	attrs = append(attrs, to...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("transfer", string(attrs))
}

// emitBurn: {"id":123,"ed":1,"ow":"ownerAddr"}
func emitBurn(id uint64, ed *uint32, owner string, _ string) {
	attrs := make([]byte, 0, len(owner)+48)
	attrs = append(attrs, '{')

	// "id":123
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// optional edition
	if ed != nil {
		attrs = append(attrs, ',', '"', 'e', 'd', '"', ':')
		attrs = strconv.AppendUint(attrs, uint64(*ed), 10)
	}

	// ,"ow":"..."
	attrs = append(attrs, ',', '"', 'o', 'w', '"', ':', '"')
	attrs = append(attrs, owner...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("burn", string(attrs))
}

// EmitCollectionCreatedEvent: {"id":5,"cr":"creatorAddr"}
func EmitCollectionCreatedEvent(id uint64, creator string) {
	attrs := make([]byte, 0, len(creator)+32)
	attrs = append(attrs, '{')

	// "id":5
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// ,"cr":"..."
	attrs = append(attrs, ',', '"', 'c', 'r', '"', ':', '"')
	attrs = append(attrs, creator...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("collection", string(attrs))
}
