#!/usr/bin/env bash
set -euo pipefail

# Mock test runner for Okinoko DAO (WASM contract)
# Author: auto-generated for tibfox, 2025-08-15
# Usage: ./mock_test.sh

echo "=========================================="
echo " Okinoko DAO Smart Contract Mock Test Run "
echo "=========================================="

# 1. Build contract to wasm
echo "[1] Building to wasm..."
GOOS=wasip1 GOARCH=wasm go build -o okinoko_dao.wasm main.go

# 2. Choose runtime (wasmtime/wasmer). Default: wasmtime
RUNTIME=${RUNTIME:-wasmtime}

if ! command -v "$RUNTIME" &>/dev/null; then
  echo "Error: $RUNTIME not installed. Install wasmtime or set RUNTIME=wasmer."
  exit 1
fi

# 3. Define sample config as JSON
CONFIG_JSON='{
  "proposal_permission":"any_member",
  "execute_permission":"any_member",
  "voting_system":"stake_based",
  "threshold_percent":60,
  "quorum_percent":40,
  "proposal_duration_secs":60,
  "execution_delay_secs":10,
  "leave_cooldown_secs":30,
  "democratic_exact_amount":1,
  "stake_min_amount":10,
  "proposal_cost":1,
  "enable_snapshot":true,
  "reward_enabled":true,
  "reward_amount":2,
  "reward_payout_on_execute":true
}'

# Helper to call contract functions
call() {
  local func="$1"; shift
  echo
  echo ">> $func $*"
  case "$RUNTIME" in
    wasmtime)
      wasmtime run --invoke "$func" ./okinoko_dao.wasm -- "$@"
      ;;
    wasmer)
      wasmer run ./okinoko_dao.wasm --invoke "$func" -- "$@"
      ;;
  esac
}

# 4. Run mock scenario
call projects_create "Okinoko Test" "Test project description" "{}" "$CONFIG_JSON" 100 "VSC"
PROJECT_ID="mock_project_id_123" # TODO: parse return value properly

call projects_get_one "$PROJECT_ID"

call projects_join "$PROJECT_ID" 20 "VSC"

call proposals_create "$PROJECT_ID" "Test Proposal" "Just testing" "{}" "bool_vote" "[\"Yes\",\"No\"]" "" 0
PROPOSAL_ID="mock_proposal_id_456"

call proposals_vote "$PROJECT_ID" "$PROPOSAL_ID" "[1]" "hash123"

call proposals_tally "$PROJECT_ID" "$PROPOSAL_ID"

call proposals_execute "$PROJECT_ID" "$PROPOSAL_ID" "VSC" || true

call projects_transfer_ownership "$PROJECT_ID" "newowner123"

call projects_pause "$PROJECT_ID" true

echo
echo "=========================================="
echo " Mock tests complete "
echo "=========================================="
