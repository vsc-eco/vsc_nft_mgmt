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

// appendEscapedJSON appends s to buf, escaping JSON special characters.
// This ensures strings containing quotes, backslashes, or control chars don't break JSON.
func appendEscapedJSON(buf []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			if c < 0x20 {
				// Other control characters: use \u00XX format
				buf = append(buf, '\\', 'u', '0', '0')
				buf = append(buf, "0123456789abcdef"[c>>4])
				buf = append(buf, "0123456789abcdef"[c&0xf])
			} else {
				buf = append(buf, c)
			}
		}
	}
	return buf
}

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
//	{"type":"mint","attributes":{"id":123,"cr":"addr","oc":"owner_col","ed":100,"nm":"name","ds":"desc","mt":"meta","st":true},"tx":"<tx>"}
func EmitMintEvent(id uint64, creator string, ownerCol string, editions uint32, name string, desc string, meta string, singleTransfer bool) {
	attrs := make([]byte, 0, len(creator)+len(ownerCol)+len(name)+len(desc)+len(meta)+64)
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

	// "nm":"name"
	attrs = append(attrs, ',', '"', 'n', 'm', '"', ':', '"')
	attrs = appendEscapedJSON(attrs, name)
	attrs = append(attrs, '"')

	// "ds":"desc"
	attrs = append(attrs, ',', '"', 'd', 's', '"', ':', '"')
	attrs = appendEscapedJSON(attrs, desc)
	attrs = append(attrs, '"')

	// "mt":"meta"
	attrs = append(attrs, ',', '"', 'm', 't', '"', ':', '"')
	attrs = appendEscapedJSON(attrs, meta)
	attrs = append(attrs, '"')

	// "st":true/false (singleTransfer)
	attrs = append(attrs, ',', '"', 's', 't', '"', ':')
	if singleTransfer {
		attrs = append(attrs, 't', 'r', 'u', 'e')
	} else {
		attrs = append(attrs, 'f', 'a', 'l', 's', 'e')
	}

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
//	{"type":"burn","attributes":{"id":123,"ed":1,"ow":"ownerAddr","col":"collectionId"},"tx":"<tx>"}
//
// Burn events are final and indicate permanent edition removal (or full NFT burn).
func emitBurn(id uint64, ed *uint32, owner string, collection string) {
	attrs := make([]byte, 0, len(owner)+len(collection)+56)
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
	attrs = append(attrs, '"')

	// "col":"collection"
	attrs = append(attrs, ',', '"', 'c', 'o', 'l', '"', ':', '"')
	attrs = append(attrs, collection...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("burn", string(attrs))
}

// ========================
// Collection Created Event
// ========================
//
// EmitCollectionCreatedEvent emits:
//
//	{"type":"collection","attributes":{"id":5,"cr":"creator","nm":"name","ds":"desc","mt":"meta"},"tx":"<tx>"}
//
// This is only called once per collection id at creation time.
func EmitCollectionCreatedEvent(id uint64, creator string, name string, desc string, meta string) {
	attrs := make([]byte, 0, len(creator)+len(name)+len(desc)+len(meta)+64)
	attrs = append(attrs, '{')

	// "id":5
	attrs = append(attrs, '"', 'i', 'd', '"', ':')
	attrs = strconv.AppendUint(attrs, id, 10)

	// "cr":"creator"
	attrs = append(attrs, ',', '"', 'c', 'r', '"', ':', '"')
	attrs = append(attrs, creator...)
	attrs = append(attrs, '"')

	// "nm":"name"
	attrs = append(attrs, ',', '"', 'n', 'm', '"', ':', '"')
	attrs = appendEscapedJSON(attrs, name)
	attrs = append(attrs, '"')

	// "ds":"desc"
	attrs = append(attrs, ',', '"', 'd', 's', '"', ':', '"')
	attrs = appendEscapedJSON(attrs, desc)
	attrs = append(attrs, '"')

	// "mt":"meta"
	attrs = append(attrs, ',', '"', 'm', 't', '"', ':', '"')
	attrs = appendEscapedJSON(attrs, meta)
	attrs = append(attrs, '"', '}')

	emitEventJSON("collection", string(attrs))
}

// ======================
// Market Contract Events
// ======================
//
// emitMarketAdd logs when a market contract is registered. Example:
//
//	{"type":"market_add","attributes":{"addr":"vscmarket"},"tx":"<tx>"}
func emitMarketAdd(addr string) {
	attrs := make([]byte, 0, len(addr)+16)
	attrs = append(attrs, '{')

	// "addr":"marketAddr"
	attrs = append(attrs, '"', 'a', 'd', 'd', 'r', '"', ':', '"')
	attrs = append(attrs, addr...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("market_add", string(attrs))
}

// emitMarketRemove logs when a market contract is removed. Example:
//
//	{"type":"market_remove","attributes":{"addr":"vscmarket"},"tx":"<tx>"}
func emitMarketRemove(addr string) {
	attrs := make([]byte, 0, len(addr)+16)
	attrs = append(attrs, '{')

	// "addr":"marketAddr"
	attrs = append(attrs, '"', 'a', 'd', 'd', 'r', '"', ':', '"')
	attrs = append(attrs, addr...)
	attrs = append(attrs, '"', '}')

	emitEventJSON("market_remove", string(attrs))
}
