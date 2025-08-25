package contract

// maintaining index keys for querying data in various ways

import (
	"encoding/json"
	"fmt"
)

// index key prefixes
const (
	idxNFTsOfCreatorPrefix      = "idx:nfts:creator:"     // + creator
	idxNFTsInCollectionPrefix   = "idx:nfts:collection:"  // + collection
	idxCollectionsOfOwnerPrefix = "idx:collection:owner:" // + owner
	idxEditionsOfGenesisNFTs    = "idx:editions:genesis:" // + genesisId
)

// addIDToIndex ensures id exists in the JSON array at indexKey (no duplicates)
func AddIDToIndex(indexKey string, id string) error {
	ptr := getStore().Get(indexKey)
	var ids []string
	if ptr != nil && *ptr != "" {
		if err := json.Unmarshal([]byte(*ptr), &ids); err != nil {
			return fmt.Errorf("unmarshal index %s: %w", indexKey, err)
		}
	}
	// check duplicate
	for _, e := range ids {
		if e == id {
			return nil
		}
	}
	ids = append(ids, id)
	b, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	getStore().Set(indexKey, string(b))
	return nil
}

// removeIDFromIndex removes id from the JSON array at indexKey (if present)
func RemoveIDFromIndex(indexKey string, id string) error {
	ptr := getStore().Get(indexKey)
	if ptr == nil || *ptr == "" {
		return nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(*ptr), &ids); err != nil {
		return fmt.Errorf("unmarshal index %s: %w", indexKey, err)
	}
	newIds := ids[:0]
	found := false
	for _, e := range ids {
		if e == id {
			found = true
			continue
		}
		newIds = append(newIds, e)
	}
	if !found {
		// nothing to do
		return nil
	}
	// if empty, store empty array "[]"
	b, err := json.Marshal(newIds)
	if err != nil {
		return err
	}
	getStore().Set(indexKey, string(b))
	return nil
}

// getIDsFromIndex returns the array of ids stored at indexKey
func GetIDsFromIndex(indexKey string) ([]string, error) {
	ptr := getStore().Get(indexKey)
	if ptr == nil || *ptr == "" {
		return []string{}, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(*ptr), &ids); err != nil {
		return nil, fmt.Errorf("unmarshal index %s: %w", indexKey, err)
	}
	return ids, nil
}
