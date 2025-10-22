package contract

import (
	"embed"
	"testing"

	"vsc-node/lib/test_utils"
	"vsc-node/modules/db/vsc/contracts"
	ledgerDb "vsc-node/modules/db/vsc/ledger"
	stateEngine "vsc-node/modules/state-processing"
)

var _ = embed.FS{} // just so "embed" can be imported

const ContractID = "vsctestcontract"
const ownerAddress = "hive:tibfox"

//go:embed artifacts/main.wasm
var ContractWasm []byte

func XTestContract(t *testing.T) {
	ct := test_utils.NewContractTest()
	ct.RegisterContract(ContractID, ownerAddress, ContractWasm)
	ct.Deposit("hive:someone", 1000, ledgerDb.AssetHive) // deposit 1 HIVE
	ct.Deposit("hive:someone", 1000, ledgerDb.AssetHbd)  // deposit 1 HBD

	ct.Call(stateEngine.TxVscCallContract{
		Self: stateEngine.TxSelf{
			TxId:                 "sometxid",
			BlockId:              "abcdef",
			Index:                69,
			OpIndex:              0,
			Timestamp:            "2025-09-03T00:00:00",
			RequiredAuths:        []string{"hive:someone"},
			RequiredPostingAuths: []string{},
		},
		ContractId: ContractID,
		Action:     "set_market",
		Payload:    []byte("asdasd"),
		RcLimit:    10000,
		Intents: []contracts.Intent{{
			Type: "transfer.allow",
			Args: map[string]string{
				"limit": "1.000",
				"token": "hive",
			},
		}},
	})
	// assert.True(t, result.Success)                 // assert contract execution success
	// assert.LessOrEqual(t, gasUsed, uint(10000000)) // assert this call uses no more than 10M WASM gas
	// assert.GreaterOrEqual(t, len(logs), 1)         // assert at least 1 log emitted
}
