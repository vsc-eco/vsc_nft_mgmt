#!/usr/bin/env bash
set -e

# -------------------------
# CONFIG
# -------------------------
SDK_MOCK_DIR="./sdk"
TEST_FILE="./contract_test.go"

mkdir -p "$SDK_MOCK_DIR"

# -------------------------
# STEP 1: Generate mock SDK
# -------------------------
cat > "$SDK_MOCK_DIR/sdk.go" <<'EOF'
package sdk

import (
	"fmt"
)

var (
	stateStore = map[string]string{}
	env        = Env{Sender: Sender{Address: Address{"addr_creator"}}}
)

type Env struct {
	Sender Sender
}
type Sender struct {
	Address Address
}
type Address struct {
	addr string
}
func (a Address) String() string { return a.addr }

func SetSender(addr string) { env.Sender.Address = Address{addr} }

func GetEnv() Env {
	return env
}
func GetEnvKey(key string) *string {
	v := ""
	return &v
}
func Log(msg string) { fmt.Println("[LOG]", msg) }

func StateSetObject(key, val string) { stateStore[key] = val }
func StateGetObject(key string) *string {
	if v, ok := stateStore[key]; ok {
		return &v
	}
	return nil
}
func StateDeleteObject(key string) {
	delete(stateStore, key)
}

func Asset(a interface{}) string {
	switch t := a.(type) {
	case string:
		return t
	default:
		return "ASSET"
	}
}
func HiveDraw(amount int64, asset string) {}
func HiveTransfer(to Address, amount int64, asset string) {
	fmt.Printf("[TRANSFER] %d %s to %s\n", amount, asset, to.String())
}
EOF

# -------------------------
# STEP 2: Generate comprehensive test file
# -------------------------
cat > "$TEST_FILE" <<'EOF'
package main

import (
	"testing"
	"okinoko_dao/sdk"
	"time"
	"encoding/json"
)

func TestFullDAOFlow(t *testing.T) {
	// Step 1: Create democratic project
	cfg := ProjectConfig{
		ProposalPermission:   PermAnyMember,
		ExecutePermission:    PermAnyMember,
		VotingSystem:         SystemDemocratic,
		ThresholdPercent:     50,
		QuorumPercent:        0,
		ProposalDurationSecs: 1,
		DemocraticExactAmt:   100,
	}
	asset := Asset{}
	pid := CreateProject("DemoProj", "desc", "{}", cfg, 100, asset)
	if pid == "" {
		t.Fatal("failed to create project")
	}

	// Step 2: Join project
	sdk.SetSender("addr_member1")
	JoinProject(pid, 100, "ASSET")
	prj := GetProject(pid)
	if len(prj.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(prj.Members))
	}

	// Step 3: Create yes/no proposal
	sdk.SetSender("addr_member1")
	proposalID, err := CreateProposal(pid, "Test Yes/No", "desc", "{}", "", TypeYesNo, nil, "addr_receiver", 50)
	if err != nil {
		t.Fatal("create proposal failed:", err)
	}

	// Step 4: Vote YES
	err = VoteProposal(pid, proposalID, []int{1}, "")
	if err != nil {
		t.Fatal("vote failed:", err)
	}

	// Step 5: Wait for duration and tally
	time.Sleep(1100 * time.Millisecond)
	passed, err := TallyProposal(pid, proposalID)
	if err != nil {
		t.Fatal("tally failed:", err)
	}
	if !passed {
		t.Fatal("proposal should have passed")
	}

	// Step 6: Execute proposal (transfer)
	err = ExecuteProposal(pid, proposalID, "ASSET")
	if err != nil {
		t.Fatal("execute failed:", err)
	}

	// Step 7: Leave project democratic refund
	sdk.SetSender("addr_member1")
	LeaveProject(pid, 0, "ASSET")

	// Step 8: Stake-based project
	cfg2 := ProjectConfig{
		ProposalPermission:   PermAnyMember,
		ExecutePermission:    PermAnyMember,
		VotingSystem:         SystemStake,
		ThresholdPercent:     50,
		QuorumPercent:        0,
		ProposalDurationSecs: 1,
		StakeMinAmt:          10,
	}
	sdk.SetSender("addr_creator2")
	pid2 := CreateProject("StakeProj", "desc", "{}", cfg2, 200, asset)
	sdk.SetSender("addr_member2")
	JoinProject(pid2, 50, "ASSET")

	// Create single-choice proposal
	sdk.SetSender("addr_member2")
	proposalID2, err := CreateProposal(pid2, "Poll", "desc", "{}", "", TypeSingle, []string{"opt1", "opt2"}, "", 0)
	if err != nil {
		t.Fatal("stake-based create proposal failed:", err)
	}

	// Vote for option 0
	err = VoteProposal(pid2, proposalID2, []int{0}, "")
	if err != nil {
		t.Fatal("vote failed:", err)
	}

	// Tally and execute
	time.Sleep(1100 * time.Millisecond)
	passed, err = TallyProposal(pid2, proposalID2)
	if err != nil {
		t.Fatal("tally failed:", err)
	}
	if !passed {
		t.Fatal("stake-based proposal should have passed")
	}
	err = ExecuteProposal(pid2, proposalID2, "ASSET")
	if err != nil {
		t.Fatal("execute stake-based proposal failed:", err)
	}

	// Step 9: Meta proposal - update_threshold
	meta := map[string]interface{}{"action": "update_threshold", "value": 75}
	metaJSON, _ := json.Marshal(meta)
	proposalID3, err := CreateProposal(pid2, "Meta Update Threshold", "desc", string(metaJSON), "", TypeMeta, nil, "", 0)
	if err != nil {
		t.Fatal("meta create failed:", err)
	}
	err = VoteProposal(pid2, proposalID3, []int{0}, "")
	if err != nil {
		t.Fatal("meta vote failed:", err)
	}
	time.Sleep(1100 * time.Millisecond)
	TallyProposal(pid2, proposalID3)
	err = ExecuteProposal(pid2, proposalID3, "ASSET")
	if err != nil {
		t.Fatal("meta execute failed:", err)
	}

	// Step 10: Meta proposal - toggle_pause explicit true
	meta = map[string]interface{}{"action": "toggle_pause", "value": true}
	metaJSON, _ = json.Marshal(meta)
	proposalID4, err := CreateProposal(pid2, "Meta Pause", "desc", string(metaJSON), "", TypeMeta, nil, "", 0)
	if err != nil {
		t.Fatal("meta toggle create failed:", err)
	}
	err = VoteProposal(pid2, proposalID4, []int{0}, "")
	if err != nil {
		t.Fatal("meta toggle vote failed:", err)
	}
	time.Sleep(1100 * time.Millisecond)
	TallyProposal(pid2, proposalID4)
	err = ExecuteProposal(pid2, proposalID4, "ASSET")
	if err != nil {
		t.Fatal("meta toggle execute failed:", err)
	}

	// Step 11: Meta proposal - toggle_pause flip
	meta = map[string]interface{}{"action": "toggle_pause"}
	metaJSON, _ = json.Marshal(meta)
	proposalID5, err := CreateProposal(pid2, "Meta Pause Flip", "desc", string(metaJSON), "", TypeMeta, nil, "", 0)
	if err != nil {
		t.Fatal("meta flip create failed:", err)
	}
	err = VoteProposal(pid2, proposalID5, []int{0}, "")
	if err != nil {
		t.Fatal("meta flip vote failed:", err)
	}
	time.Sleep(1100 * time.Millisecond)
	TallyProposal(pid2, proposalID5)
	err = ExecuteProposal(pid2, proposalID5, "ASSET")
	if err != nil {
		t.Fatal("meta flip execute failed:", err)
	}
}
EOF

# -------------------------
# STEP 3: Run tests
# -------------------------
echo "[*] Running tests..."
go test -v
