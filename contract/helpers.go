package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// Conversions from/to json strings

func ToJSON[T any](v T, objectType string) string {
	b, err := json.Marshal(v)
	if err != nil {
		sdk.Abort("failed to marshal " + objectType)
	}
	return string(b)
}

func FromJSON[T any](data string, objectType string) *T {
	// data = strings.TrimSpace(data)
	var v T
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		sdk.Abort(
			fmt.Sprintf("failed to unmarshal %s \ninput: %s\nerror: %v", objectType, data, err.Error()))
	}
	return &v
}

func StringToUInt64(ptr *string) uint64 {
	if ptr == nil {
		sdk.Abort("input is empty")
	}
	val, err := strconv.ParseUint(*ptr, 10, 64) // base 10, 64-bit
	if err != nil {
		sdk.Abort(fmt.Sprintf("failed to parse '%s' to uint64: %w", *ptr, err))
	}
	return val
}

func UInt64ToString(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func UInt64ArrayToJsonString(ids []uint64, objectHint string) *string {
	jsonStr := ToJSON(ids, objectHint)
	return &jsonStr
}
