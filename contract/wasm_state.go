// wasm_state.go
package main

import "okinoko_dao/sdk"

type WasmState struct{}

func (WasmState) Set(key, value string) {
	sdk.StateSetObject(key, value)
}

func (WasmState) Get(key string) *string {
	return sdk.StateGetObject(key)
}

func (WasmState) Delete(key string) {
	sdk.StateDeleteObject(key)
}
