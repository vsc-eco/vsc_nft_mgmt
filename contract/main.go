package main

// ========================================
// Global Contract Constants & Base Helpers
// ========================================
//
// This file holds core constants and lightweight utility functions

const (
	maxNameLength = 48                   // upper bound for NFT or collection names
	maxDescLength = 128                  // max allowed chars for description field
	maxMetaLength = 32768                // max metadata size (32KB) - generous limit to prevent state bloat
	contractOwner = "hive:contractowner" // contractOwner can add/remove supported market contract
)

func main() {
	// placeholder function
}
