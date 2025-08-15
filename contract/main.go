////////////////////////////////////////////////////////////////////////////////
// Okinoko DAO: A universal DAO for the vsc network
// created by tibfox 2025-08-12
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"okinoko_dao/sdk"
	"strconv"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// Types & Structs
////////////////////////////////////////////////////////////////////////////////

// Voting system for a project
type VotingSystem string

const (
	SystemDemocratic VotingSystem = "democratic"  // every member has an equal vote
	SystemStake      VotingSystem = "stake_based" // ever member has a different vote weight - based on the stake in the project treasury fund
)

// Proposal types
type VotingType string

const (
	VotingTypeBoolVote     VotingType = "bool_vote"     // proposals with boolean vote - these can also execute transfers
	VotingTypeSingleChoice VotingType = "single_choice" // proposals with only one answer as vote
	VotingTypeMultiChoice  VotingType = "multi_choice"  // proposals with multiple possible answers as vote
	VotingTypeMetaProposal VotingType = "meta"          // meta proposals to change project settings
)

// Permission for who may create/execute proposals
type Permission string

const (
	PermCreatorOnly Permission = "creator_only"
	PermAnyMember   Permission = "any_member"
	PermAnyone      Permission = "anyone"
)

// Proposal state lifecycle
type ProposalState string

const (
	StateActive     ProposalState = "active"     // default state for new proposals
	StateExecutable ProposalState = "executable" // quorum is reached - final
	StateExecuted   ProposalState = "executed"   // proposal passed
	StateFailed     ProposalState = "failed"     // proposal failed to gather enough votes within the proposal duration
)

// Role constants
const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Member represents a project member
type Member struct {
	Address       string `json:"address"`
	Stake         int64  `json:"stake"`
	Role          string `json:"role"`           // "admin" or "member"
	JoinedAt      int64  `json:"joined_at"`      // unix ts
	LastActionAt  int64  `json:"last_action_at"` // last stake/join/withdraw time (for cooldown)
	ExitRequested int64  `json:"exit_requested"` // 0 if not requested
	Reputation    int64  `json:"reputation"`     // initially 0 | every vote += 1 | every passed proposal += 5
}

// ProjectConfig contains toggles & params for a project
type ProjectConfig struct {
	ProposalPermission    Permission   `json:"proposal_permission"`      // who may create proposals
	ExecutePermission     Permission   `json:"execute_permission"`       // who may execute transfers
	VotingSystem          VotingSystem `json:"voting_system"`            // democratic or stake_based
	ThresholdPercent      int          `json:"threshold_percent"`        // percent required to pass (0-100)
	QuorumPercent         int          `json:"quorum_percent"`           // percent of voting power that must participate (0-100)
	ProposalDurationSecs  int64        `json:"proposal_duration_secs"`   // default duration
	ExecutionDelaySecs    int64        `json:"execution_delay_secs"`     // delay after pass before exec allowed
	LeaveCooldownSecs     int64        `json:"leave_cooldown_secs"`      // cooldown for leaving/withdrawing
	DemocraticExactAmt    int64        `json:"democratic_exact_amount"`  // exact amount required to join democratic
	StakeMinAmt           int64        `json:"stake_min_amount"`         // min stake for stake-based joining
	ProposalCost          int64        `json:"proposal_cost"`            // fee for creating proposals (goes to project funds)
	EnableSnapshot        bool         `json:"enable_snapshot"`          // snapshot member stakes at proposal start
	RewardEnabled         bool         `json:"reward_enabled"`           // rewards enabled
	RewardAmount          int64        `json:"reward_amount"`            // reward for proposer (from funds)
	RewardPayoutOnExecute bool         `json:"reward_payout_on_execute"` // pay reward when proposal executed
}

// Project - stored under project:<id>
type Project struct {
	ID           string            `json:"id"`
	Owner        string            `json:"owner"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	JsonMetadata string            `json:"json_metadata"`
	Config       ProjectConfig     `json:"config"`
	Members      map[string]Member `json:"members"` // key: address string
	Funds        int64             `json:"funds"`   // pool in minimal unit
	FundsAsset   sdk.Asset         `json:"funds_asset"`
	CreatedAt    int64             `json:"created_at"`
	Paused       bool              `json:"paused"`
}

// Proposal - stored separately at proposal:<id>
type Proposal struct {
	ID              string        `json:"id"`
	ProjectID       string        `json:"project_id"`
	Creator         string        `json:"creator"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	JsonMetadata    string        `json:"json_metadata"`
	Type            VotingType    `json:"type"`
	Options         []string      `json:"options"` // for polls
	Receiver        string        `json:"receiver,omitempty"`
	Amount          int64         `json:"amount,omitempty"`
	CreatedAt       int64         `json:"created_at"`
	DurationSeconds int64         `json:"duration_seconds"` // if 0 => use project default
	State           ProposalState `json:"state"`
	Passed          bool          `json:"passed"`
	Executed        bool          `json:"executed"`
	FinalizedAt     int64         `json:"finalized_at,omitempty"`
	PassTimestamp   int64         `json:"pass_timestamp,omitempty"`
	SnapshotTotal   int64         `json:"snapshot_total,omitempty"` // total voting power snapshot if enabled
	TxID            string        `json:"tx_id,omitempty"`
}

type VoteRecord struct {
	ProjectID   string `json:"project_id"`
	ProposalID  string `json:"proposal_id"`
	Voter       string `json:"voter"`
	ChoiceIndex []int  `json:"choice_index"` // indexes for options; for yes/no -> [0] or [1]
	Weight      int64  `json:"weight"`
	VotedAt     int64  `json:"voted_at"`
}

////////////////////////////////////////////////////////////////////////////////
// Helpers: keys, guids, time
////////////////////////////////////////////////////////////////////////////////

func getSenderAddress() string {
	return sdk.GetEnv().Sender.Address.String()
}

func projectKey(id string) string {
	return "project:" + id
}

func projectProposalsIndexKey(projectID string) string {
	return "project:" + projectID + ":proposals"
}

func proposalKey(id string) string {
	return "proposal:" + id
}

func voteKey(projectID, proposalID, voter string) string {
	return fmt.Sprintf("vote:%s:%s:%s", projectID, proposalID, voter)
}

const projectsIndexKey = "projects:index"

// generateGUID returns a 16-byte hex string
func generateGUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("g_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func nowUnix() int64 {
	// try chain timestamp via env key
	if tsPtr := sdk.GetEnvKey("block.timestamp"); tsPtr != nil && *tsPtr != "" {
		// try parse as integer seconds
		if v, err := strconv.ParseInt(*tsPtr, 10, 64); err == nil {
			return v
		}
		// try RFC3339
		if t, err := time.Parse(time.RFC3339, *tsPtr); err == nil {
			return t.Unix()
		}
	}
	return time.Now().Unix()
}

func getTxID() string {
	if t := sdk.GetEnvKey("tx.id"); t != nil {
		return *t
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// Contract State Persistence helpers
////////////////////////////////////////////////////////////////////////////////

func saveProject(pro *Project) {
	key := projectKey(pro.ID)
	b, _ := json.Marshal(pro)
	sdk.StateSetObject(key, string(b))
}

func loadProject(id string) (*Project, error) {
	key := projectKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil {
		return nil, fmt.Errorf("project %s not found", id)
	}
	var pro Project
	if err := json.Unmarshal([]byte(*ptr), &pro); err != nil {
		return nil, fmt.Errorf("failed unmarshal project %s: %v", id, err)
	}
	return &pro, nil
}

func addProjectToIndex(id string) {
	ptr := sdk.StateGetObject(projectsIndexKey)
	var ids []string
	if ptr != nil {
		json.Unmarshal([]byte(*ptr), &ids)
	}
	// prevent duplicates
	for _, v := range ids {
		if v == id {
			return
		}
	}
	ids = append(ids, id)
	b, _ := json.Marshal(ids)
	sdk.StateSetObject(projectsIndexKey, string(b))
}

func listAllProjectIDs() []string {
	ptr := sdk.StateGetObject(projectsIndexKey)
	if ptr == nil {
		return []string{}
	}
	var ids []string
	if err := json.Unmarshal([]byte(*ptr), &ids); err != nil {
		return []string{}
	}
	return ids
}

func saveProposal(prpsl *Proposal) {
	key := proposalKey(prpsl.ID)
	b, _ := json.Marshal(prpsl)
	sdk.StateSetObject(key, string(b))
}

func loadProposal(id string) (*Proposal, error) {
	key := proposalKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil {
		return nil, fmt.Errorf("proposal %s not found", id)
	}
	var prpsl Proposal
	if err := json.Unmarshal([]byte(*ptr), &prpsl); err != nil {
		return nil, fmt.Errorf("failed unmarshal proposal %s: %v", id, err)
	}
	return &prpsl, nil
}

func addProposalToProjectIndex(projectID, proposalID string) {
	key := projectProposalsIndexKey(projectID)
	ptr := sdk.StateGetObject(key)
	var ids []string
	if ptr != nil {
		json.Unmarshal([]byte(*ptr), &ids)
	}
	// TODO: avoid dublicates
	ids = append(ids, proposalID)
	b, _ := json.Marshal(ids)
	sdk.StateSetObject(key, string(b))
}

func listProposalIDsForProject(projectID string) []string {
	key := projectProposalsIndexKey(projectID)
	ptr := sdk.StateGetObject(key)
	if ptr == nil {
		return []string{}
	}
	var ids []string
	if err := json.Unmarshal([]byte(*ptr), &ids); err != nil {
		return []string{}
	}
	return ids
}

func saveVote(vote *VoteRecord) {
	key := voteKey(vote.ProjectID, vote.ProposalID, vote.Voter)
	b, _ := json.Marshal(vote)
	sdk.StateSetObject(key, string(b))

	// ensure voter listed in index for iteration (store list under project:proposal:voters)
	votersKey := fmt.Sprintf("proposal:%s:%s:voters", vote.ProjectID, vote.ProposalID)
	ptr := sdk.StateGetObject(votersKey)
	var voters []string
	if ptr != nil {
		json.Unmarshal([]byte(*ptr), &voters)
	}
	seen := false
	for _, a := range voters {
		if a == vote.Voter {
			seen = true
			break
		}
	}
	if !seen {
		voters = append(voters, vote.Voter)
		nb, _ := json.Marshal(voters)
		sdk.StateSetObject(votersKey, string(nb))
	}
}

func loadVotesForProposal(projectID, proposalID string) []VoteRecord {
	votersKey := fmt.Sprintf("proposal:%s:%s:voters", projectID, proposalID)
	ptr := sdk.StateGetObject(votersKey)
	if ptr == nil {
		return []VoteRecord{}
	}
	var voters []string
	if err := json.Unmarshal([]byte(*ptr), &voters); err != nil {
		return []VoteRecord{}
	}
	out := make([]VoteRecord, 0, len(voters))
	for _, v := range voters {
		vk := voteKey(projectID, proposalID, v)
		vp := sdk.StateGetObject(vk)
		if vp == nil {
			continue
		}
		var vr VoteRecord
		if err := json.Unmarshal([]byte(*vp), &vr); err == nil {
			out = append(out, vr)
		}
	}
	return out
}

// remove vote only needed if member leaves project while still voted on an active proposal
func removeVote(projectID, proposalID, voter string) {
	key := voteKey(projectID, proposalID, voter)
	sdk.StateDeleteObject(key)
	// remove from voter list
	votersKey := fmt.Sprintf("proposal:%s:%s:voters", projectID, proposalID)
	ptr := sdk.StateGetObject(votersKey)
	if ptr == nil {
		return
	}
	var voters []string
	json.Unmarshal([]byte(*ptr), &voters)
	newV := make([]string, 0, len(voters))
	for _, a := range voters {
		if a != voter {
			newV = append(newV, a)
		}
	}
	nb, _ := json.Marshal(newV)
	sdk.StateSetObject(votersKey, string(nb))
}

////////////////////////////////////////////////////////////////////////////////
// Public contract functions
////////////////////////////////////////////////////////////////////////////////

//go:wasmexport projects_create
func CreateProject(name, description, jsonMetadata string, cfg ProjectConfig, amount int64, asset sdk.Asset) string {
	if amount <= 0 {
		sdk.Log("CreateProject: amount must be > 1")
		return "XXX" // TODO: correct return
	}
	// TODO: add more asset checks here

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
		Config:       cfg,
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
	// if it is stake based - add stake of the project fee as stake
	if cfg.VotingSystem == SystemStake {
		m.Stake = amount
	}

	prj.Members[creator] = m
	saveProject(&prj)
	addProjectToIndex(id)
	sdk.Log("CreateProject: " + id)
	return id
}

// GetProject - returns the project object (no proposals included)
//
//go:wasmexport projects_get_one
func GetProject(projectID string) *Project {
	prj, err := loadProject(projectID)
	if err != nil {
		return nil
	}
	return prj
}

// GetAllProjects - returns all projects (IDs then loads each)
//
//go:wasmexport projects_get_all
func GetAllProjects() []*Project {
	ids := listAllProjectIDs()
	out := make([]*Project, 0, len(ids))
	for _, id := range ids {
		if prj, err := loadProject(id); err == nil {
			out = append(out, prj)
		}
	}
	return out
}

// AddFunds - draw funds from caller and add to project's treasury pool
// If the project is a stake based system & the sender is a valid mamber then the stake of the member will get updated accordingly.

//go:wasmexport projects_add_funds
func AddFunds(projectID string, amount int64, asset string) {
	if amount <= 0 {
		sdk.Log("AddFunds: amount must be > 0")
		return
	}
	prj, err := loadProject(projectID)
	if err != nil {
		sdk.Log("AddFunds: project not found")
		return
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
}

//go:wasmexport projects_join
func JoinProject(projectID string, amount int64, assetString string) {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	asset := sdk.Asset(assetString)

	if err != nil {
		sdk.Log("JoinProject: project not found")
		return
	}
	if prj.Paused {
		sdk.Log("JoinProject: project paused")
		return
	}
	if amount <= 0 {
		sdk.Log("JoinProject: amount must be > 0")
		return
	}
	if asset != prj.FundsAsset {
		sdk.Log(fmt.Sprintf("JoinProject: asset must match the project main asset: %s", prj.FundsAsset.String()))
		return
	}

	now := nowUnix()
	if prj.Config.VotingSystem == SystemDemocratic {
		if amount != prj.Config.DemocraticExactAmt {
			sdk.Log(fmt.Sprintf("JoinProject: democratic projects need an exact amount to join: %d %s", prj.Config.DemocraticExactAmt, prj.FundsAsset.String()))
			return
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

			return
		}
		_, ok := prj.Members[caller]
		if ok {
			sdk.Log("JoinProject: already member")
			return
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
}

//go:wasmexport projects_leave
func LeaveProject(projectID string, withdrawAmount int64, asset string) {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		sdk.Log("LeaveProject: project not found")
		return
	}
	if prj.Paused {
		sdk.Log("LeaveProject: project paused")
		return
	}
	member, ok := prj.Members[caller]
	if !ok {
		sdk.Log("LeaveProject: not a member")
		return
	}

	now := nowUnix()
	// if exit requested previously -> try to withdraw
	if member.ExitRequested > 0 {
		if now-member.ExitRequested < prj.Config.LeaveCooldownSecs {
			sdk.Log("LeaveProject: cooldown not passed")
			return
		}
		// withdraw stake (for stake-based) or refund democratic amount
		if prj.Config.VotingSystem == SystemDemocratic {
			refund := prj.Config.DemocraticExactAmt
			if prj.Funds < refund {
				sdk.Log("LeaveProject: insufficient project funds")
				return
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
			return
		}
		// stake-based
		withdraw := member.Stake
		if withdraw <= 0 {
			sdk.Log("LeaveProject: nothing to withdraw")
			return
		}
		if prj.Funds < withdraw {
			sdk.Log("LeaveProject: insufficient project funds")
			return
		}
		prj.Funds -= withdraw
		sdk.HiveTransfer(sdk.Address(caller), withdraw, sdk.Asset(asset))
		delete(prj.Members, caller)
		for _, pid := range listProposalIDsForProject(projectID) {
			removeVote(projectID, pid, caller)
		}
		saveProject(prj)
		sdk.Log("LeaveProject: withdrew stake")
		return
	}

	// otherwise set exit requested timestamp
	member.ExitRequested = now
	prj.Members[caller] = member
	saveProject(prj)
	sdk.Log("LeaveProject: exit requested")
}

//go:wasmexport proposals_create
func CreateProposal(projectID string, name string, description string, jsonMetadata string,
	vtype VotingType, options []string, receiver string, amount int64) (string, error) {

	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		return "", err
	}
	if prj.Paused {
		return "", fmt.Errorf("project paused")
	}

	// permission check
	if prj.Config.ProposalPermission == PermCreatorOnly && caller != prj.Owner {
		return "", fmt.Errorf("only project owner can create proposals")
	}
	if prj.Config.ProposalPermission == PermAnyMember {
		if _, ok := prj.Members[caller]; !ok {
			return "", fmt.Errorf("only members can create proposals")
		}
	}

	// options validation
	if (vtype == VotingTypeSingleChoice || vtype == VotingTypeMultiChoice) && len(options) == 0 {
		return "", fmt.Errorf("poll proposals require options")
	}

	// charge proposal cost (draw funds from caller to contract)
	if prj.Config.ProposalCost > 0 {
		sdk.HiveDraw(prj.Config.ProposalCost, sdk.Asset("VSC")) // TODO: what happens when not enough funds?!
		prj.Funds += prj.Config.ProposalCost
		saveProject(prj)
	}

	// create proposal
	id := generateGUID()
	now := nowUnix()
	duration := prj.Config.ProposalDurationSecs
	if duration <= 0 {
		duration = 60 * 60 * 24 * 7 // default 7 days
	}
	prpsl := Proposal{
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

	// If snapshot enabled, compute snapshot total voting power
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

	saveProposal(&prpsl)
	addProposalToProjectIndex(projectID, id)
	sdk.Log("CreateProposal: " + id + " in project " + projectID)
	return id, nil
}

//go:wasmexport proposals_vote
func VoteProposal(projectID, proposalID string, choices []int, commitHash string) error {
	caller := getSenderAddress()
	now := nowUnix()

	prj, err := loadProject(projectID)
	if err != nil {
		return err
	}
	if prj.Paused {
		return fmt.Errorf("project paused")
	}
	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return err
	}
	if prpsl.State != StateActive {
		return fmt.Errorf("proposal not active")
	}
	// time window
	if prpsl.DurationSeconds > 0 && now > prpsl.CreatedAt+prpsl.DurationSeconds {
		// finalize instead
		TallyProposal(projectID, proposalID)
		return fmt.Errorf("voting period ended (tallied)")
	}
	// member check
	member, ok := prj.Members[caller]
	if !ok {
		return fmt.Errorf("only members may vote")
	}
	// compute weight
	var weight int64 = 1
	if prj.Config.VotingSystem == SystemStake {
		weight = member.Stake
		if weight <= 0 {
			return fmt.Errorf("member stake zero")
		}
	}

	// validate choices by type
	switch prpsl.Type {
	case VotingTypeBoolVote:
		if len(choices) != 1 || (choices[0] != 0 && choices[0] != 1) {
			return fmt.Errorf("yes_no requires single choice 0 or 1")
		}
	case VotingTypeSingleChoice:
		if len(choices) != 1 {
			return fmt.Errorf("single_choice requires exactly 1 index")
		}
		if choices[0] < 0 || choices[0] >= len(prpsl.Options) {
			return fmt.Errorf("option index out of range")
		}
	case VotingTypeMultiChoice:
		if len(choices) == 0 {
			return fmt.Errorf("multi_choice requires >=1 choices")
		}
		for _, idx := range choices {
			if idx < 0 || idx >= len(prpsl.Options) {
				return fmt.Errorf("option index out of range")
			}
		}
	case VotingTypeMetaProposal:
		// same validations as polls depending on meta semantics
	default:
		return fmt.Errorf("unknown proposal type")
	}

	vote := VoteRecord{
		ProjectID:   projectID,
		ProposalID:  proposalID,
		Voter:       caller,
		ChoiceIndex: choices,
		Weight:      weight,

		VotedAt: now,
	}

	saveVote(&vote)
	sdk.Log("VoteProposal: voter " + caller + " for " + proposalID)
	return nil
}

//go:wasmexport proposals_tally
func TallyProposal(projectID, proposalID string) (bool, error) {
	prj, err := loadProject(projectID)
	if err != nil {
		return false, err
	}
	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return false, err
	}

	// Only tally if proposal duration has passed
	if nowUnix() < prpsl.CreatedAt+prj.Config.ProposalDurationSecs {
		// Proposal duration not yet over, do not change state
		sdk.Log("TallyProposal: proposal duration not over yet")
		return false, nil
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
		return false, nil
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
	return prpsl.Passed, nil
}

//go:wasmexport proposals_execute
func ExecuteProposal(projectID, proposalID, asset string) error {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		return err
	}
	if prj.Paused {
		return fmt.Errorf("project paused")
	}
	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return err
	}
	if prpsl.State != StateExecutable {
		return fmt.Errorf("proposal not executable (state=%s)", prpsl.State)
	}
	// check execution delay
	if prj.Config.ExecutionDelaySecs > 0 && nowUnix() < prpsl.PassTimestamp+prj.Config.ExecutionDelaySecs {
		return fmt.Errorf("execution delay not passed")
	}
	// check permission to execute
	if prj.Config.ExecutePermission == PermCreatorOnly && caller != prpsl.Creator {
		return fmt.Errorf("only creator can execute")
	}
	if prj.Config.ExecutePermission == PermAnyMember {
		if _, ok := prj.Members[caller]; !ok {
			return fmt.Errorf("only members can execute")
		}
	}
	// For transfers: only yes/no allowed
	if prpsl.Type == VotingTypeBoolVote && prpsl.Amount > 0 && strings.TrimSpace(prpsl.Receiver) != "" {
		// ensure funds
		if prj.Funds < prpsl.Amount {
			return fmt.Errorf("insufficient project funds")
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
		return nil
	}
	// Meta proposals: interpret json_metadata to perform allowed changes
	if prpsl.Type == VotingTypeMetaProposal {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(prpsl.JsonMetadata), &meta); err != nil {
			return fmt.Errorf("invalid meta json")
		}
		action, _ := meta["action"].(string)
		switch action {
		// TODO: add more project properties here
		case "update_threshold":
			if v, ok := meta["value"].(float64); ok {
				newv := int(v)
				if newv < 0 || newv > 100 {
					return fmt.Errorf("threshold out of range")
				}
				prj.Config.ThresholdPercent = newv
				prpsl.Executed = true
				prpsl.State = StateExecuted
				prpsl.FinalizedAt = nowUnix()
				saveProject(prj)
				saveProposal(prpsl)
				sdk.Log("ExecuteProposal: updated threshold")
				return nil
			}
		case "toggle_pause":
			if val, ok := meta["value"].(bool); ok {
				prj.Paused = val
				prpsl.Executed = true
				prpsl.State = StateExecuted
				prpsl.FinalizedAt = nowUnix()
				saveProject(prj)
				saveProposal(prpsl)
				sdk.Log("ExecuteProposal: toggled pause")
				return nil
			} else {
				// flip
				prj.Paused = !prj.Paused
				prpsl.Executed = true
				prpsl.State = StateExecuted
				prpsl.FinalizedAt = nowUnix()
				saveProject(prj)
				saveProposal(prpsl)
				sdk.Log("ExecuteProposal: toggled pause (flip)")
				return nil
			}
		// TODO: add more meta actions here (update quorum, proposal cost, reward setting, etc.)
		default:
			return fmt.Errorf("unknown meta action")
		}
	}
	// If nothing to execute, mark executed
	prpsl.Executed = true
	prpsl.State = StateExecuted
	prpsl.FinalizedAt = nowUnix()
	saveProposal(prpsl)
	sdk.Log("ExecuteProposal: marked executed without transfer " + proposalID)
	return nil
}

//go:wasmexport proposals_get_one
func GetProposal(proposalID string) *Proposal {
	prpsl, err := loadProposal(proposalID)
	if err != nil {
		return nil
	}
	return prpsl
}

//go:wasmexport proposals_get_all
func GetProjectProposals(projectID string) []Proposal {
	ids := listProposalIDsForProject(projectID)
	out := make([]Proposal, 0, len(ids))
	for _, id := range ids {
		if prpsl, err := loadProposal(id); err == nil {
			out = append(out, *prpsl)
		}
	}
	return out
}

//go:wasmexport projects_transfer_ownership
func TransferProjectOwnership(projectID, newOwner string) error {
	caller := getSenderAddress()

	prj, err := loadProject(projectID)
	if err != nil {
		return err
	}
	if caller != prj.Owner {
		return fmt.Errorf("only creator")
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
	return nil
}

//go:wasmexport projects_pause
func EmergencyPauseImmediate(projectID string, pause bool) error {
	caller := getSenderAddress()
	prj, err := loadProject(projectID)
	if err != nil {
		return err
	}
	if caller != prj.Owner {
		return fmt.Errorf("only the project owner can pause / unpause without dedicated meta proposal")
	}
	prj.Paused = pause
	saveProject(prj)
	sdk.Log("EmergencyPauseImmediate: set paused=" + strconv.FormatBool(pause))
	return nil
}

// TODO: removing projects only via meta proposals?
// // RemoveProject - deletes project and all its proposals and votes (creator only)
// // projects_remove
// func RemoveProject(projectID string) error {
// 	env := sdk.GetEnv()
// 	caller := env.Sender.Address.String()

// 	prj, err := loadProject(projectID)
// 	if err != nil {
// 		return err
// 	}
// 	if caller != prj.Owner {
// 		return fmt.Errorf("only owner can remove")
// 	}
// 	// remove index entry
// 	ids := listAllProjectIDs()
// 	newIds := make([]string, 0, len(ids))
// 	for _, id := range ids {
// 		if id != projectID {
// 			newIds = append(newIds, id)
// 		}
// 	}
// 	nb, _ := json.Marshal(newIds)
// 	sdk.StateSetObject(projectsIndexKey, string(nb))

// 	// delete proposals & votes
// 	for _, pid := range listProposalIDsForProject(projectID) {
// 		// delete votes
// 		votersKey := fmt.Sprintf("proposal:%s:%s:voters", projectID, pid)
// 		ptr := sdk.StateGetObject(votersKey)
// 		if ptr != nil {
// 			var voters []string
// 			json.Unmarshal([]byte(*ptr), &voters)
// 			for _, v := range voters {
// 				sdk.StateDeleteObject(voteKey(projectID, pid, v))
// 			}
// 			sdk.StateDeleteObject(votersKey)
// 		}
// 		// delete proposal
// 		sdk.StateDeleteObject(proposalKey(pid))
// 	}
// 	// delete proposals index
// 	sdk.StateDeleteObject(projectProposalsIndexKey(projectID))
//  	// TODO: refund project members

// 	// delete project
// 	sdk.StateDeleteObject(projectKey(projectID))
// 	sdk.Log("RemoveProject: removed " + projectID)
// 	return nil
// }
