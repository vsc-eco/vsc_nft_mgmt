package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// Index functions
// for fast and more effective reads we maintain indexes in the contract state
// all indexes are split into chunks of X entries to avoid overflowing the max size of a key/value of the contract state

// index key prefixes
const (
	maxChunkSize           = 2500
	NFTsCreator            = "n:crtr"     // + creator		// holds nfts minted by a given user (only unique and genesis editions)
	CollectionsOwner       = "c:ownr:"    // + owner			// holds collections for a given user (to avoid duplicate names)
	NFTsCollection         = "n:c:"       // + collection		// holds nfts contained in a given collection
	AllEditionsOfGenesis   = "e_all:g:"   // + genesisId		// holds editions for a given genesis edition
	AvailEditionsOfGenesis = "e_avail:g:" // + genesisId		// holds available editions for a given genesis edition
	NFTsCount              = "cnt:n"      // 					// holds a int counter for nfts (to create new ids)
	CollectionCount        = "cnt:c"      // 					// holds a int counter for collections (to create new ids)
)

// stores number of chunks for a base index
func chunkCounterKey(base string) string {
	return base + ":ch"
}

func chunkKey(base string, chunk int) string {
	return base + ":" + strconv.Itoa(chunk)
}

// get number of chunks for an index
func getChunkCount(baseKey string) int {
	ptr := sdk.StateGetObject(chunkCounterKey(baseKey))
	if ptr == nil || *ptr == "" {
		return 0
	}
	n, _ := strconv.Atoi(*ptr)
	return n
}

// set number of chunks
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

// ensures id exists across all chunks (no duplicates).
func AddIDToIndex(baseKey string, id uint64) {
	chunks := getChunkCount(baseKey)

	// search existing chunks for duplicates or free space
	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)

		ids := []uint64{}
		if ptr != nil && *ptr != "" {
			ids = *FromJSON[[]uint64](*ptr, "index "+key)

			// duplicate check
			for _, e := range ids {
				if e == id {
					return // already present
				}
			}

			// append if space
			if len(ids) < maxChunkSize {
				ids = append(ids, id)
				sdk.StateSetObject(key, ToJSON(ids, "index "+key))
				return
			}
		}
	}

	// not found / no space -> create new chunk
	key := chunkKey(baseKey, chunks)
	ids := []uint64{id}
	sdk.StateSetObject(key, ToJSON(ids, "index "+key))
	setChunkCount(baseKey, chunks+1)
}

// removes id from whichever chunk it’s in.
func RemoveIDFromIndex(baseKey string, id uint64) {
	chunks := getChunkCount(baseKey)
	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)
		if ptr == nil || *ptr == "" {
			continue
		}

		ids := *FromJSON[[]uint64](*ptr, "index "+key)
		newIds := ids[:0]
		found := false

		for _, e := range ids {
			if e == id {
				found = true
				continue
			}
			newIds = append(newIds, e)
		}

		if found {
			sdk.StateSetObject(key, ToJSON(newIds, "index "+key))
		}
	}
}

// collects all IDs across all chunks.
func GetIDsFromIndex(baseKey string) []uint64 {
	all := []uint64{}
	chunks := getChunkCount(baseKey)

	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)
		if ptr == nil || *ptr == "" {
			continue
		}

		ids := *FromJSON[[]uint64](*ptr, "index "+key)
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
	idSet := make(map[uint64]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	// iterate existing chunks
	for i := 0; i < chunks; i++ {
		key := chunkKey(baseKey, i)
		ptr := sdk.StateGetObject(key)

		var chunkIds []uint64
		if ptr != nil && *ptr != "" {
			chunkIds = *FromJSON[[]uint64](*ptr, "index "+key)

			// build a set of existing IDs to avoid duplicates
			existing := make(map[uint64]struct{}, len(chunkIds))
			for _, e := range chunkIds {
				existing[e] = struct{}{}
			}

			// append as many IDs as fit
			for id := range idSet {
				if len(chunkIds) >= maxChunkSize {
					break
				}
				if _, ok := existing[id]; !ok {
					chunkIds = append(chunkIds, id)
					delete(idSet, id)
				}
			}

			sdk.StateSetObject(key, ToJSON(chunkIds, "index "+key))
			if len(idSet) == 0 {
				return // all IDs added
			}
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
		sdk.StateSetObject(key, ToJSON(chunkIds, "index "+key))
		chunks++
	}
	setChunkCount(baseKey, chunks)
}
