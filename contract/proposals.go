package main

import (
	"encoding/json"
	"okinoko_dao/sdk"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------
// Create Proposal (returns full proposal as JSON)
// Exports only primitive params; slices come in as JSON strings
// -----------------------------------------------------------------------------
//
//go:wasmexport proposals_create
func CreateProposal(
	projectID string,
	name string,
	description string,
	jsonMetadata string,
	vtypeStr string, // VotingType as string (e.g., "bool_vote", "single_choice", ...)
	optionsJSON string, // JSON array of strings, e.g. '["opt1","opt2"]'
	receiver string,
	amount int64,
) string {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"proposals_create", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if prj.Paused {
		return returnJsonResponse(
			"proposals_create", false, map[string]interface{}{
				"details": "project is paused",
			},
		)
	}

	// Permission check
	if prj.Config.ProposalPermission == PermCreatorOnly && caller != prj.Owner {
		return returnJsonResponse(
			"proposals_create", false, map[string]interface{}{
				"details": "only project owner can create proposals",
			},
		)

	}
	if prj.Config.ProposalPermission == PermAnyMember {
		if _, ok := prj.Members[caller]; !ok {
			return returnJsonResponse(
				"proposals_create", false, map[string]interface{}{
					"details": "only members can create proposals",
				},
			)

		}
	}

	// Parse VotingType
	vtype := VotingType(vtypeStr)
	switch vtype {
	case VotingTypeBoolVote, VotingTypeSingleChoice, VotingTypeMultiChoice, VotingTypeMetaProposal:
	default:
		return returnJsonResponse(
			"proposals_create", false, map[string]interface{}{
				"details": "invalid voting type",
			},
		)
	}

	// Parse options JSON (for polls)
	var options []string
	if strings.TrimSpace(optionsJSON) != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
			return returnJsonResponse(
				"proposals_create", false, map[string]interface{}{
					"details": "invalid options json",
				},
			)

		}
	}
	// Options validation
	if (vtype == VotingTypeSingleChoice || vtype == VotingTypeMultiChoice) && len(options) == 0 {
		return returnJsonResponse(
			"proposals_create", false, map[string]interface{}{
				"details": "poll proposals require options",
			},
		)
	}

	// Charge proposal cost (from caller to contract)
	if prj.Config.ProposalCost > 0 {
		sdk.HiveDraw(prj.Config.ProposalCost, sdk.Asset("VSC")) // NOTE: same placeholder as your original code
		prj.Funds += prj.Config.ProposalCost
		saveProject(prj)
	}

	// Create proposal
	id := generateGUID()
	now := nowUnix()
	duration := prj.Config.ProposalDurationSecs
	if duration <= 0 {
		duration = 60 * 60 * 24 * 7 // default 7 days
	}
	prpsl := &Proposal{
		ID:              id,
		ProjectID:       projectID,
		Creator:         caller,
		Name:            name,
		Description:     description,
		JsonMetadata:    jsonMetadata,
		Type:            vtype,
		Options:         options,
		Receiver:        receiver,
		Amount:          amount,
		CreatedAt:       now,
		DurationSeconds: duration,
		State:           StateActive,
		Passed:          false,
		Executed:        false,
		TxID:            getTxID(),
	}

	// Snapshot voting power (if enabled)
	if prj.Config.EnableSnapshot {
		var total int64 = 0
		if prj.Config.VotingSystem == SystemDemocratic {
			total = int64(len(prj.Members))
		} else {
			for _, m := range prj.Members {
				total += m.Stake
			}
		}
		prpsl.SnapshotTotal = total
	}

	saveProposal(prpsl)
	addProposalToProjectIndex(projectID, id)
	sdk.Log("CreateProposal: " + id + " in project " + projectID)

	return returnJsonResponse(
		"proposals_create", true, map[string]interface{}{
			"id":       id,
			"proposal": prpsl,
		},
	)

}

// -----------------------------------------------------------------------------
// Vote Proposal (choices as JSON array) -> returns {"success":true} or {"error":...}
// -----------------------------------------------------------------------------
//
//go:wasmexport proposals_vote
func VoteProposal(projectID, proposalID string, choicesJSON string, commitHash string) string {
	caller := getSenderAddress()
	now := nowUnix()

	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if prj.Paused {
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "project is paused",
			},
		)
	}

	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "proposal not found",
			},
		)
	}
	if prpsl.State != StateActive {
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "proposal is not active",
			},
		)
	}

	// Time window
	if prpsl.DurationSeconds > 0 && now > prpsl.CreatedAt+prpsl.DurationSeconds {
		// finalize instead
		_ = TallyProposal(projectID, proposalID) // ignore result here
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "voting period ended (tallied)",
			},
		)
	}

	// Member check
	member, ok := prj.Members[caller]
	if !ok {
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "only members may vote",
			},
		)
	}

	// Compute weight
	weight := int64(1)
	if prj.Config.VotingSystem == SystemStake {
		weight = member.Stake
		if weight <= 0 {
			return returnJsonResponse(
				"proposals_vote", false, map[string]interface{}{
					"details": "member stake zero",
				},
			)
		}
	}

	// Parse choices
	var choices []int
	if err := json.Unmarshal([]byte(choicesJSON), &choices); err != nil {
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "invalid choice json",
			},
		)
	}

	// Validate choices by type
	switch prpsl.Type {
	case VotingTypeBoolVote:
		if len(choices) != 1 || (choices[0] != 0 && choices[0] != 1) {
			return returnJsonResponse(
				"proposals_vote", false, map[string]interface{}{
					"details": "bool_vote requires single choice 0 or 1",
				},
			)

		}
	case VotingTypeSingleChoice:
		if len(choices) != 1 {
			return returnJsonResponse(
				"proposals_vote", false, map[string]interface{}{
					"details": "single_choice requires exactly 1 index",
				},
			)

		}
		if choices[0] < 0 || choices[0] >= len(prpsl.Options) {

			return returnJsonResponse(
				"proposals_vote", false, map[string]interface{}{
					"details": "option index out of range",
				},
			)

		}
	case VotingTypeMultiChoice:
		if len(choices) == 0 {
			return returnJsonResponse(
				"proposals_vote", false, map[string]interface{}{
					"details": "multi_choice requires >=1 choices",
				},
			)
		}
		for _, idx := range choices {
			if idx < 0 || idx >= len(prpsl.Options) {
				return returnJsonResponse(
					"proposals_vote", false, map[string]interface{}{
						"details": "option index out of range",
					},
				)

			}
		}
	case VotingTypeMetaProposal:
		// no extra validation here
	default:
		return returnJsonResponse(
			"proposals_vote", false, map[string]interface{}{
				"details": "unknown proposal type",
			},
		)
	}

	vote := VoteRecord{
		ProjectID:   projectID,
		ProposalID:  proposalID,
		Voter:       caller,
		ChoiceIndex: choices,
		Weight:      weight,
		VotedAt:     now,
	}

	saveVote(&vote)
	sdk.Log("VoteProposal: voter " + caller + " for " + proposalID)
	return returnJsonResponse(
		"proposals_vote", true, map[string]interface{}{
			"vote": vote,
		},
	)
}

// -----------------------------------------------------------------------------
// Tally Proposal -> returns {"success":true,"passed":bool,"proposal":{...}}
// -----------------------------------------------------------------------------
//
//go:wasmexport proposals_tally
func TallyProposal(projectID, proposalID string) string {
	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"proposals_tally", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return returnJsonResponse(
			"proposals_tally", false, map[string]interface{}{
				"details": "proposal not found",
			},
		)
	}

	// Use proposal-specific duration if set, else project default
	duration := prpsl.DurationSeconds
	if duration <= 0 {
		duration = prj.Config.ProposalDurationSecs
	}
	// Only tally if proposal duration has passed
	if nowUnix() < prpsl.CreatedAt+duration {
		return returnJsonResponse(
			"proposals_tally", false, map[string]interface{}{
				"details": "proposal duration not over yet",
			},
		)
	}

	// compute total possible voting power
	var totalPossible int64 = 0
	if prj.Config.VotingSystem == SystemDemocratic {
		totalPossible = int64(len(prj.Members))
	} else {
		for _, m := range prj.Members {
			totalPossible += m.Stake
		}
	}
	if prj.Config.EnableSnapshot && prpsl.SnapshotTotal > 0 {
		totalPossible = prpsl.SnapshotTotal
	}

	// gather votes
	votes := loadVotesForProposal(projectID, proposalID)
	// compute participation and option counts
	var participation int64 = 0
	optionCounts := make(map[int]int64)
	for _, v := range votes {
		participation += v.Weight
		for _, idx := range v.ChoiceIndex {
			optionCounts[idx] += v.Weight
		}
	}

	// check quorum
	required := int64(0)
	if prj.Config.QuorumPercent > 0 {
		required = (int64(prj.Config.QuorumPercent)*totalPossible + 99) / 100 // ceil
	}
	if required > 0 && participation < required {
		prpsl.Passed = false
		prpsl.State = StateFailed
		prpsl.FinalizedAt = nowUnix()
		saveProposal(prpsl)
		sdk.Log("TallyProposal: quorum not reached")
		return returnJsonResponse(
			"proposals_tally", false, map[string]interface{}{
				"details": "quorum not reached",
			},
		)
	}

	// evaluate result by type
	switch prpsl.Type {
	case VotingTypeBoolVote:
		yes := optionCounts[1]
		if yes*100 >= int64(prj.Config.ThresholdPercent)*totalPossible {
			prpsl.Passed = true
			prpsl.State = StateExecutable
			prpsl.PassTimestamp = nowUnix()
		} else {
			prpsl.Passed = false
			prpsl.State = StateFailed
		}
	case VotingTypeSingleChoice, VotingTypeMultiChoice, VotingTypeMetaProposal:
		// find best option weight
		bestIdx := -1
		var bestVal int64 = 0
		for idx, cnt := range optionCounts {
			if cnt > bestVal {
				bestVal = cnt
				bestIdx = idx
			}
		}
		if bestIdx >= 0 && bestVal*100 >= int64(prj.Config.ThresholdPercent)*totalPossible {
			prpsl.Passed = true
			prpsl.State = StateExecutable
			prpsl.PassTimestamp = nowUnix()
		} else {
			prpsl.Passed = false
			prpsl.State = StateFailed
		}
	default:
		prpsl.Passed = false
		prpsl.State = StateFailed
	}

	prpsl.FinalizedAt = nowUnix()
	saveProposal(prpsl)
	sdk.Log("TallyProposal: tallied - passed=" + strconv.FormatBool(prpsl.Passed))
	return returnJsonResponse(
		"proposals_tally", true, map[string]interface{}{
			"passed":   prpsl.Passed,
			"proposal": prpsl,
		},
	)

}

// -----------------------------------------------------------------------------
// Execute Proposal -> returns JSON with action details or error
// (keeps your original flow; only return type changed)
// -----------------------------------------------------------------------------
//
//go:wasmexport proposals_execute
func ExecuteProposal(projectID, proposalID, asset string) string {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"proposals_execute", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if prj.Paused {
		return returnJsonResponse(
			"proposals_execute", false, map[string]interface{}{
				"details": "project is paused",
			},
		)
	}
	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return returnJsonResponse(
			"proposals_execute", false, map[string]interface{}{
				"details": "proposal not found",
			},
		)
	}
	if prpsl.State != StateExecutable {
		return returnJsonResponse(
			"proposals_execute", false, map[string]interface{}{
				"details":  "proposal not executable",
				"proposal": prpsl,
			},
		)

	}
	// check execution delay
	if prj.Config.ExecutionDelaySecs > 0 && nowUnix() < prpsl.PassTimestamp+prj.Config.ExecutionDelaySecs {
		return returnJsonResponse(
			"proposals_execute", false, map[string]interface{}{
				"details": "execution delay not passed",
				"delay":   prj.Config.ExecutionDelaySecs,
				//TODO: add current elapsed time
			},
		)
	}
	// check permission to execute
	if prj.Config.ExecutePermission == PermCreatorOnly && caller != prpsl.Creator {
		return returnJsonResponse(
			"proposals_execute", false, map[string]interface{}{
				"details": "only creator can execute",
				"creator": prpsl.Creator,
			},
		)

	}
	if prj.Config.ExecutePermission == PermAnyMember {
		if _, ok := prj.Members[caller]; !ok {
			return returnJsonResponse(
				"proposals_execute", false, map[string]interface{}{
					"details": "only members can execute",
				},
			)
		}
	}

	// For transfers: only yes/no allowed
	if prpsl.Type == VotingTypeBoolVote && prpsl.Amount > 0 && strings.TrimSpace(prpsl.Receiver) != "" {
		// ensure funds
		if prj.Funds < prpsl.Amount {
			return returnJsonResponse(
				"proposals_execute", false, map[string]interface{}{
					"details": "insufficient project funds",
					"funds":   prj.Funds,
					"needed":  prpsl.Amount,
					"asset":   prj.FundsAsset,
				},
			)

		}
		// transfer from contract to receiver
		sdk.HiveTransfer(sdk.Address(prpsl.Receiver), prpsl.Amount, sdk.Asset(asset))
		prj.Funds -= prpsl.Amount
		prpsl.Executed = true
		prpsl.State = StateExecuted
		prpsl.FinalizedAt = nowUnix()
		saveProposal(prpsl)
		saveProject(prj)
		// reward proposer if enabled
		if prj.Config.RewardEnabled && prj.Config.RewardPayoutOnExecute && prj.Config.RewardAmount > 0 && prj.Funds >= prj.Config.RewardAmount {
			prj.Funds -= prj.Config.RewardAmount
			// transfer reward to proposer
			sdk.HiveTransfer(sdk.Address(prpsl.Creator), prj.Config.RewardAmount, sdk.Asset(asset))
			saveProject(prj)
		}
		sdk.Log("ExecuteProposal: transfer executed " + proposalID)
		return returnJsonResponse(
			"proposals_execute", true, map[string]interface{}{
				"to":     prpsl.Receiver,
				"amount": prpsl.Amount,
				"asset":  asset,
			},
		)

	}

	// Meta proposals: interpret json_metadata to perform allowed changes
	if prpsl.Type == VotingTypeMetaProposal {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(prpsl.JsonMetadata), &meta); err != nil {
			return returnJsonResponse(
				"proposals_execute", false, map[string]interface{}{
					"details": "invalid meta json",
				},
			)

		}
		action, _ := meta["action"].(string)
		switch action {
		// TODO: add more project properties here
		case "update_threshold":
			if v, ok := meta["value"].(float64); ok {
				newv := int(v)
				if newv < 0 || newv > 100 {
					return returnJsonResponse(
						"proposals_execute", false, map[string]interface{}{
							"property": "update_threshold",
							"details":  "value out of range",
						},
					)

				}
				prj.Config.ThresholdPercent = newv
				prpsl.Executed = true
				prpsl.State = StateExecuted
				prpsl.FinalizedAt = nowUnix()
				saveProject(prj)
				saveProposal(prpsl)
				sdk.Log("ExecuteProposal: updated threshold")
				return returnJsonResponse(
					"proposals_execute", true, map[string]interface{}{
						"property": "update_threshold",
						"value":    newv,
					},
				)

			}
		case "toggle_pause":
			if val, ok := meta["value"].(bool); ok {
				prj.Paused = val
			} else {
				// flip
				prj.Paused = !prj.Paused
			}
			prpsl.Executed = true
			prpsl.State = StateExecuted
			prpsl.FinalizedAt = nowUnix()
			saveProject(prj)
			saveProposal(prpsl)
			sdk.Log("ExecuteProposal: toggled pause")
			return returnJsonResponse(
				"proposals_execute", true, map[string]interface{}{
					"property": "toggle_pause",
					"value":    prj.Paused,
				},
			)
		// TODO: add more meta actions here (update quorum, proposal cost, reward setting, etc.)
		default:
			return returnJsonResponse(
				"proposals_execute", false, map[string]interface{}{
					"details": "meta property unknown",
				},
			)
		}
	}

	// If nothing to execute, mark executed
	prpsl.Executed = true
	prpsl.State = StateExecuted
	prpsl.FinalizedAt = nowUnix()
	saveProposal(prpsl)
	sdk.Log("ExecuteProposal: marked executed without transfer " + proposalID)
	return returnJsonResponse(
		"proposals_execute", true, map[string]interface{}{
			"details": "executed without meta change or transfer",
		},
	)
}

// -----------------------------------------------------------------------------
// Get single proposal -> returns {"success":true,"proposal":{...}}
// -----------------------------------------------------------------------------
//
//go:wasmexport proposals_get_one
func GetProposal(proposalID string) string {
	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return returnJsonResponse(
			"proposals_get_one", false, map[string]interface{}{
				"details": "proposal not found",
			},
		)
	}
	return returnJsonResponse(
		"proposals_get_one", true, map[string]interface{}{
			"propsal": prpsl,
		},
	)
}

// -----------------------------------------------------------------------------
// Get all proposals for a project -> returns {"success":true,"proposals":[...]}
// -----------------------------------------------------------------------------
//
//go:wasmexport proposals_get_all
func GetProjectProposals(projectID string) string {
	ids := listProposalIDsForProject(projectID)
	proposals := make([]Proposal, 0, len(ids))
	for _, id := range ids {
		if prpsl, err := loadProposal(id); err == nil {
			proposals = append(proposals, *prpsl)
		}
	}
	return returnJsonResponse(
		"proposals_get_one", true, map[string]interface{}{
			"propsal": proposals,
		},
	)
}
