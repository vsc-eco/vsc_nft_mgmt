package main

import (
	"vsc_nft_mgmt/sdk"
)

const (
	NFTsCount       = "cnt:n" //                  // holds a int counter for nfts (to create new ids)
	CollectionCount = "cnt:c" //                  // holds a int counter for collections (to create new ids)
)

// ---- helpers ----

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
