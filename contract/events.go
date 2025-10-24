package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

//
// ==========================================
// Centralized Event Emission (Gas-Optimized)
// ==========================================
//
// We use manual JSON construction via byte buffers to minimize allocations,
// avoid reflection, and guarantee deterministic output. Events are logged via
// sdk.Log and are intended for indexing off-chain behavior.
//

// emitEventJSON emits a generic JSON-formatted event.
// `attributesJSON` must already be a valid JSON object such as {"id":1,"cr":"alice"}.
// This function wraps it with type and transaction metadata:
//
//	{"type":"mint","attributes":{...},"tx":"<tx-id>"}
var (
	keyType       = []byte(`"type":"`)
	keyAttributes = []byte(`","attributes":`)
	keyTx         = []byte(`,"tx":"`)
	braceClose    = []byte{'"', '}'}
)

func emitEventJSON(eventType string, attributesJSON string) {
	txID := sdk.GetEnvKey("tx.id")

	// pre-allocate the buffer
	b := make([]byte, 0, len(eventType)+len(attributesJSON)+len(*txID)+32)
	b = append(b, '{')
	b = append(b, keyType...)
	b = append(b, eventType...)
	b = append(b, keyAttributes...)
	b = append(b, attributesJSON...)
	b = append(b, keyTx...)
	b = append(b, (*txID)...)
	b = append(b, braceClose...)

	sdk.Log(string(b))
}

// ==============
// NFT Mint Event
// ==============
//
// EmitMintEvent formats and emits:
//
//	{"type":"mint","attributes":{"id":123,"cr":"addr","oc":"owner_col","ed":100},"tx":"<tx>"}
func EmitMintEvent(id uint64, creator string, ownerCol string, editions uint32) {
	attrs := make([]byte, 0, len(creator)+len(ownerCol)+32)
	attrs = append(attrs, '{')

	// "id":123
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// "cr":"creator"
	attrs = append(attrs, ',', '"', 'c', 'r', '"', ':', '"')
	attrs = append(attrs, creator...)
	attrs = append(attrs, '"')

	// "oc":"owner_collection"
	attrs = append(attrs, ',', '"', 'o', 'c', '"', ':', '"')
	attrs = append(attrs, ownerCol...)
	attrs = append(attrs, '"')

	// "ed":<number-of-editions>
	attrs = append(attrs, ',', '"', 'e', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, uint64(editions), 10)

	attrs = append(attrs, '}')
	emitEventJSON("mint", string(attrs))
}

// ==================
// NFT Transfer Event
// ==================
//
// emitTransfer logs ownership change. Example:
//
//	{"type":"transfer","attributes":{"id":123,"ed":1,"fr":"fromAddr","to":"toAddr"},"tx":"<tx>"}
func emitTransfer(id uint64, ed *uint32, from string, to string) {
	attrs := make([]byte, 0, len(from)+len(to)+48)
	attrs = append(attrs, '{')

	// "id":123
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// Optional edition
	if ed != nil {
		attrs = append(attrs, ',', '"', 'e', 'd', '"', ':')
		attrs = strconv.AppendUint(attrs, uint64(*ed), 10)
	}

	// "fr":"fromAddr"
	attrs = append(attrs, ',', '"', 'f', 'r', '"', ':', '"')
	attrs = append(attrs, from...)
	attrs = append(attrs, '"')

	// "to":"toAddr"
	attrs = append(attrs, ',', '"', 't', 'o', '"', ':', '"')
	attrs = append(attrs, to...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("transfer", string(attrs))
}

// ==============
// NFT Burn Event
// ==============
//
// emitBurn logs a burn event for NFT or specific edition. Example:
//
//	{"type":"burn","attributes":{"id":123,"ed":1,"ow":"ownerAddr"},"tx":"<tx>"}
//
// Burn events are final and indicate permanent edition removal (or full NFT burn).
func emitBurn(id uint64, ed *uint32, owner string, _ string) {
	attrs := make([]byte, 0, len(owner)+48)
	attrs = append(attrs, '{')

	// "id":123
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// Optional edition
	if ed != nil {
		attrs = append(attrs, ',', '"', 'e', 'd', '"', ':')
		attrs = strconv.AppendUint(attrs, uint64(*ed), 10)
	}

	// "ow":"owner"
	attrs = append(attrs, ',', '"', 'o', 'w', '"', ':', '"')
	attrs = append(attrs, owner...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("burn", string(attrs))
}

// ========================
// Collection Created Event
// ========================
//
// EmitCollectionCreatedEvent emits:
//
//	{"type":"collection","attributes":{"id":5,"cr":"creator"},"tx":"<tx>"}
//
// This is only called once per collection id at creation time.
func EmitCollectionCreatedEvent(id uint64, creator string) {
	attrs := make([]byte, 0, len(creator)+32)
	attrs = append(attrs, '{')

	// "id":5
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// "cr":"creator"
	attrs = append(attrs, ',', '"', 'c', 'r', '"', ':', '"')
	attrs = append(attrs, creator...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("collection", string(attrs))
}
