package main

import (
	"strconv"
	"strings"
	"vsc_nft_mgmt/sdk"
)

// Index functions
// For fast and more effective reads we maintain indexes in the contract state.
// All indexes are split into chunks of X entries to avoid overflowing the max size of a key/value of the contract state.

// index key prefixes
const (
	maxChunkSize           = 2500
	NFTsCreator            = "n:crtr"     // + creator        // holds nfts minted by a given user (only unique and genesis editions)
	CollectionsOwner       = "c:ownr:"    // + owner          // holds collections for a given user (to avoid duplicate names)
	NFTsCollection         = "n:c:"       // + collection     // holds nfts contained in a given collection
	AllEditionsOfGenesis   = "e_all:g:"   // + genesisId      // holds editions for a given genesis edition
	AvailEditionsOfGenesis = "e_avail:g:" // + genesisId      // holds available editions for a given genesis edition
	NFTsCount              = "cnt:n"      //                  // holds a int counter for nfts (to create new ids)
	CollectionCount        = "cnt:c"      //                  // holds a int counter for collections (to create new ids)
)

// ---- helpers ----

func chunkCounterKey(base string) string {
	return base + ":ch"
}

func chunkKey(base string, chunk int) string {
	return base + ":" + strconv.Itoa(chunk)
}

func getChunkCount(baseKey string) int {
	ptr := sdk.StateGetObject(chunkCounterKey(baseKey))
	if ptr == nil || *ptr == "" {
		return 0
	}
	n, _ := strconv.Atoi(*ptr)
	return n
}

func setChunkCount(baseKey string, n int) {
	sdk.StateSetObject(chunkCounterKey(baseKey), strconv.Itoa(n))
}

func getCount(key string) uint64 {
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		return 0
	}
	return StringToUInt64(ptr)
}

func setCount(key string, n uint64) {
	sdk.StateSetObject(key, UInt64ToString(n))
}

// ---- chunk <-> slice conversion ----

func parseIDs(s string) []uint64 {
	if s == "" {
		return []uint64{}
	}
	parts := strings.Split(s, ",")
	ids := make([]uint64, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		ids = append(ids, StringToUInt64(&p))
	}
	return ids
}

func formatIDs(ids []uint64) string {
	if len(ids) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i, id := range ids {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(UInt64ToString(id))
	}
	return sb.String()
}

// ---- index operations ----

// AddIDToIndex ensures id exists across all chunks (no duplicates).
func AddIDToIndex(baseKey string, id uint64) {
	chunks := getChunkCount(baseKey)

	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)
		if ptr == nil {
			continue
		}
		ids := parseIDs(*ptr)

		// O(1) duplicate check
		idMap := make(map[uint64]struct{}, len(ids))
		for _, e := range ids {
			idMap[e] = struct{}{}
		}
		if _, exists := idMap[id]; exists {
			return
		}

		// append if space
		if len(ids) < maxChunkSize {
			ids = append(ids, id)
			sdk.StateSetObject(key, formatIDs(ids))
			return
		}
	}

	// not found / no space -> create new chunk
	key := chunkKey(baseKey, chunks)
	sdk.StateSetObject(key, formatIDs([]uint64{id}))
	setChunkCount(baseKey, chunks+1)
}

// RemoveIDFromIndex removes id from whichever chunk it’s in (swap-and-trim method).
func RemoveIDFromIndex(baseKey string, id uint64) {
	chunks := getChunkCount(baseKey)
	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)
		if ptr == nil || *ptr == "" {
			continue
		}
		ids := parseIDs(*ptr)
		found := false
		for j := 0; j < len(ids); j++ {
			if ids[j] == id {
				// swap with last element and trim slice
				ids[j] = ids[len(ids)-1]
				ids = ids[:len(ids)-1]
				found = true
				break
			}
		}
		if found {
			sdk.StateSetObject(key, formatIDs(ids))
			return // stop after first removal
		}
	}
}

// GetIDsFromIndex collects all IDs across all chunks (preallocate slice for efficiency).
func GetIDsFromIndex(baseKey string) []uint64 {
	chunks := getChunkCount(baseKey)
	all := make([]uint64, 0, chunks*maxChunkSize)
	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)
		if ptr == nil || *ptr == "" {
			continue
		}
		ids := parseIDs(*ptr)
		all = append(all, ids...)
	}
	return all
}

// SliceToIndex adds a slice of IDs to a given baseKey, filling existing chunks first,
// then creating new chunks as needed. Avoids duplicates.
func SliceToIndex(baseKey string, ids []uint64) {
	if len(ids) == 0 {
		return
	}

	chunks := getChunkCount(baseKey)
	idSet := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	// iterate existing chunks
	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)
		if ptr == nil {
			continue
		}
		chunkIds := parseIDs(*ptr)
		added := false

		// remove already existing IDs
		for _, e := range chunkIds {
			delete(idSet, e)
		}

		// append new IDs to fill the chunk
		for id := range idSet {
			if len(chunkIds) >= maxChunkSize {
				break
			}
			chunkIds = append(chunkIds, id)
			delete(idSet, id)
			added = true
		}

		if added {
			sdk.StateSetObject(key, formatIDs(chunkIds))
		}
		if len(idSet) == 0 {
			return
		}
	}

	// add remaining IDs in new chunks
	for len(idSet) > 0 {
		key := chunkKey(baseKey, chunks)
		chunkIds := make([]uint64, 0, maxChunkSize)
		for id := range idSet {
			if len(chunkIds) >= maxChunkSize {
				break
			}
			chunkIds = append(chunkIds, id)
			delete(idSet, id)
		}
		sdk.StateSetObject(key, formatIDs(chunkIds))
		chunks++
	}
	setChunkCount(baseKey, chunks)
}
