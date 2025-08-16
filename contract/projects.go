package main

import (
	"fmt"
	"okinoko_dao/sdk"
	"strconv"
)

//go:wasmexport projects_create
func CreateProject(name, description, jsonMetadata string, cfgJSON string, amount int64, assetJSON string) string {
	if amount <= 0 {
		sdk.Log("CreateProject: amount must be > 1")
		return returnJsonResponse(
			"CreatePrprojects_createoject", false, map[string]interface{}{
				"details": "mount must be > 1",
			},
		)

	}

	// Parse JSON params
	cfg, err := ProjectConfigFromJSON(cfgJSON)
	if err != nil {
		return returnJsonResponse(
			"projects_create", false, map[string]interface{}{
				"details": "invalid project config",
			},
		)

	}
	var asset sdk.Asset
	if err := FromJSON(assetJSON, &asset); err != nil {
		return returnJsonResponse(
			"projects_create", false, map[string]interface{}{
				"details": "invalid asset",
			},
		)

	}

	creator := getSenderAddress()

	sdk.HiveDraw(amount, sdk.Asset(asset))

	id := generateGUID()
	now := nowUnix()

	prj := Project{
		ID:           id,
		Owner:        creator,
		Name:         name,
		Description:  description,
		JsonMetadata: jsonMetadata,
		Config:       *cfg,
		Members:      map[string]Member{},
		Funds:        amount,
		FundsAsset:   asset,
		CreatedAt:    now,
		Paused:       false,
	}

	// Add creator as admin
	m := Member{
		Address:      creator,
		Stake:        1,
		Role:         RoleAdmin,
		JoinedAt:     now,
		LastActionAt: now,
		Reputation:   0,
	}
	// If it is stake-based, add full stake
	if cfg.VotingSystem == SystemStake {
		m.Stake = amount
	}
	prj.Members[creator] = m

	saveProject(&prj)
	addProjectToIndex(id)

	sdk.Log("CreateProject: " + id)

	return returnJsonResponse(
		"projects_create", true, map[string]interface{}{
			"id":      id,
			"project": prj,
		},
	)

}

// GetProject - returns the project object as JSON
//
//go:wasmexport projects_get_one
func GetProject(projectID string) string {
	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"projects_get_one", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	return returnJsonResponse(
		"projects_get_one", true, map[string]interface{}{
			"project": prj,
		},
	)
}

// GetAllProjects - returns all projects as JSON array
//
//go:wasmexport projects_get_all
func GetAllProjects() string {
	ids := listAllProjectIDs()
	projects := make([]*Project, 0, len(ids))
	for _, id := range ids {
		if prj, err := loadProject(id); err == nil {
			projects = append(projects, prj)
		}
	}
	return returnJsonResponse(
		"projects_get_all", true, map[string]interface{}{
			"projects": projects,
		},
	)
}

// AddFunds - draw funds from caller and add to project's treasury pool
// If the project is a stake based system & the sender is a valid mamber then the stake of the member will get updated accordingly.
//
//go:wasmexport projects_add_funds
func AddFunds(projectID string, amount int64, asset string) string {
	if amount <= 0 {

		return returnJsonResponse(
			"projects_add_funds", false, map[string]interface{}{
				"projects": "amount needs to be > 0",
			},
		)
	}
	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"projects_add_funds", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if asset != prj.FundsAsset.String() {
		return returnJsonResponse(
			"projects_add_funds", false, map[string]interface{}{
				"details": "asset needs to be " + prj.FundsAsset.String(),
			},
		)
	}
	caller := getSenderAddress()

	sdk.HiveDraw(amount, sdk.Asset(asset))
	prj.Funds += amount

	// if stake based
	if prj.Config.VotingSystem == SystemStake {
		// check if member
		m, ismember := prj.Members[caller]
		if ismember {
			now := nowUnix()
			m.Stake = m.Stake + amount
			m.LastActionAt = now
			// add member with exact stake
			prj.Members[caller] = m
		}
	}

	saveProject(prj)
	sdk.Log("AddFunds: added " + strconv.FormatInt(amount, 10))
	return returnJsonResponse(
		"projects_add_funds", true, map[string]interface{}{
			"added": amount,
			"asset": prj.FundsAsset.String(),
		},
	)
}

//go:wasmexport projects_join
func JoinProject(projectID string, amount int64, assetString string) string {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	asset := sdk.Asset(assetString)

	if err != nil {
		return returnJsonResponse(
			"projects_join", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if prj.Paused {
		return returnJsonResponse(
			"projects_join", false, map[string]interface{}{
				"details": "project is paused",
			},
		)
	}
	if amount <= 0 {
		return returnJsonResponse(
			"projects_join", false, map[string]interface{}{
				"details": "amount needs to be > 0",
			},
		)
	}
	if asset != prj.FundsAsset {
		sdk.Log(fmt.Sprintf("JoinProject: asset must match the project main asset: %s", prj.FundsAsset.String()))
		return returnJsonResponse(
			"projects_join", false, map[string]interface{}{
				"details": "asset needs to be " + prj.FundsAsset.String(),
			},
		)
	}

	now := nowUnix()
	if prj.Config.VotingSystem == SystemDemocratic {
		if amount != prj.Config.DemocraticExactAmt {
			sdk.Log(fmt.Sprintf("JoinProject: democratic projects need an exact amount to join: %d %s", prj.Config.DemocraticExactAmt, prj.FundsAsset.String()))
			return returnJsonResponse(
				"projects_join", false, map[string]interface{}{
					"details": fmt.Sprintf("democratic projects need an exact amount to join: %d %s", prj.Config.DemocraticExactAmt, prj.FundsAsset.String()),
				},
			)
		}
		// transfer funds into contract
		sdk.HiveDraw(amount, sdk.Asset(asset)) // TODO: what if not enough funds?!

		// add member with stake 1
		prj.Members[caller] = Member{
			Address:      caller,
			Stake:        1,
			Role:         RoleMember,
			JoinedAt:     now,
			LastActionAt: now,
			Reputation:   0,
		}
		prj.Funds += amount
	} else { // if the project is a stake based system
		if amount < prj.Config.StakeMinAmt {
			sdk.Log(fmt.Sprintf("JoinProject: the sent amount < than the minimum projects entry fee: %d %s", prj.Config.StakeMinAmt, prj.FundsAsset.String()))

			return returnJsonResponse(
				"projects_join", false, map[string]interface{}{
					"details": fmt.Sprintf("the sent amount < than the minimum projects entry fee: %d %s", prj.Config.StakeMinAmt, prj.FundsAsset.String()),
				},
			)
		}
		_, ok := prj.Members[caller]
		if ok {
			sdk.Log("JoinProject: already member")
			return returnJsonResponse(
				"projects_join", false, map[string]interface{}{
					"details": "already member",
				},
			)
		} else {
			// transfer funds into contract
			sdk.HiveDraw(amount, sdk.Asset(asset)) // TODO: what if not enough funds?!
			// add member with exact stake
			prj.Members[caller] = Member{
				Address:      caller,
				Stake:        amount,
				Role:         RoleMember,
				JoinedAt:     now,
				LastActionAt: now,
			}
		}
		prj.Funds += amount
	}
	saveProject(prj)
	sdk.Log("JoinProject: " + projectID + " by " + caller)
	return returnJsonResponse(
		"projects_join", true, map[string]interface{}{
			"joined": caller,
		},
	)
}

//go:wasmexport projects_leave
func LeaveProject(projectID string, withdrawAmount int64, asset string) string {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"projects_leave", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if prj.Paused {
		sdk.Log("LeaveProject: project paused")
		return returnJsonResponse(
			"projects_leave", false, map[string]interface{}{
				"details": "project is paused",
			},
		)
	}
	member, ok := prj.Members[caller]
	if !ok {
		sdk.Log("LeaveProject: not a member")
		return returnJsonResponse(
			"projects_leave", false, map[string]interface{}{
				"details": caller + " is not a member",
			},
		)
	}

	now := nowUnix()
	// if exit requested previously -> try to withdraw
	if member.ExitRequested > 0 {
		if now-member.ExitRequested < prj.Config.LeaveCooldownSecs {
			sdk.Log("LeaveProject: cooldown not passed")
			return returnJsonResponse(
				"projects_leave", false, map[string]interface{}{
					"details": "cooldown of " + caller + " is not passed yet",
					// TODO add remaining time
				},
			)
		}
		// withdraw stake (for stake-based) or refund democratic amount
		if prj.Config.VotingSystem == SystemDemocratic {
			refund := prj.Config.DemocraticExactAmt
			if prj.Funds < refund {
				sdk.Log("LeaveProject: insufficient project funds")
				return returnJsonResponse(
					"projects_leave", false, map[string]interface{}{
						"details": "insufficient project funds",
					},
				)
			}
			prj.Funds -= refund
			// transfer back to caller
			sdk.HiveTransfer(sdk.Address(caller), refund, sdk.Asset(asset))
			delete(prj.Members, caller)
			// remove votes
			for _, pid := range listProposalIDsForProject(projectID) {
				removeVote(projectID, pid, caller)
			}
			saveProject(prj)
			sdk.Log("LeaveProject: democratic refunded")
			return returnJsonResponse(
				"projects_leave", true, map[string]interface{}{
					"details": "democratic refunded",
				},
			)
		}
		// stake-based
		withdraw := member.Stake
		if withdraw <= 0 { // should not happen(?)
			sdk.Log("LeaveProject: nothing to withdraw")
			return returnJsonResponse(
				"projects_leave", true, map[string]interface{}{
					"details": "nothing to withdraw",
				},
			)
		}
		if prj.Funds < withdraw {
			sdk.Log("LeaveProject: insufficient project funds")
			return returnJsonResponse(
				"projects_leave", false, map[string]interface{}{
					"details": "insufficient project funds",
				},
			)
		}
		prj.Funds -= withdraw
		sdk.HiveTransfer(sdk.Address(caller), withdraw, sdk.Asset(asset))
		delete(prj.Members, caller)
		for _, pid := range listProposalIDsForProject(projectID) {
			removeVote(projectID, pid, caller)
		}
		saveProject(prj)
		sdk.Log("LeaveProject: withdrew stake")
		return returnJsonResponse(
			"projects_leave", true, map[string]interface{}{
				"details": "stake refunded",
			},
		)
	}

	// otherwise set exit requested timestamp
	member.ExitRequested = now
	prj.Members[caller] = member
	saveProject(prj)
	return returnJsonResponse(
		"projects_leave", true, map[string]interface{}{
			"details": "exit requested.",
			//TODO: add cooldown info
		},
	)
}

//go:wasmexport projects_transfer_ownership
func TransferProjectOwnership(projectID, newOwner string) string {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"projects_transfer_ownership", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if caller != prj.Owner {
		return returnJsonResponse(
			"projects_transfer_ownership", false, map[string]interface{}{
				"details": caller + " is not owner of the project",
			},
		)
	}
	prj.Owner = newOwner
	// ensure new owner exists as member
	if _, ok := prj.Members[newOwner]; !ok {
		prj.Members[newOwner] = Member{
			Address:      newOwner,
			Stake:        0,
			Role:         RoleMember,
			JoinedAt:     nowUnix(),
			LastActionAt: nowUnix(),
		}
	}
	saveProject(prj)
	sdk.Log("TransferProjectOwnership: " + projectID + " -> " + newOwner)
	return returnJsonResponse(
		"projects_transfer_ownership", true, map[string]interface{}{
			"to": newOwner,
		},
	)

}

//go:wasmexport projects_pause
func EmergencyPauseImmediate(projectID string, pause bool) string {
	caller := getSenderAddress()
	prj, err := loadProject(projectID)
	if err != nil {
		return returnJsonResponse(
			"projects_pause", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}
	if caller != prj.Owner {
		return returnJsonResponse(
			"projects_pause", false, map[string]interface{}{
				"details": "only the project owner can pause / unpause without dedicated meta proposal",
			},
		)

	}
	prj.Paused = pause
	saveProject(prj)
	sdk.Log("EmergencyPauseImmediate: set paused=" + strconv.FormatBool(pause))
	return returnJsonResponse(
		"projects_pause", true, map[string]interface{}{
			"details": "pause switched",
			"value":   pause,
		},
	)
}
