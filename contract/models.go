package main

import (
	"okinoko_dao/sdk"
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

func (m *Member) ToJSON() (string, error) {
	return ToJSON(m)
}
func MemberFromJSON(data string) (*Member, error) {
	var m Member
	if err := FromJSON(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
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

func (pc *ProjectConfig) ToJSON() (string, error) {
	return ToJSON(pc)
}
func ProjectConfigFromJSON(data string) (*ProjectConfig, error) {
	var pc ProjectConfig
	if err := FromJSON(data, &pc); err != nil {
		return nil, err
	}
	return &pc, nil
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

func (p *Project) ToJSON() (string, error) {
	return ToJSON(p)
}
func ProjectFromJSON(data string) (*Project, error) {
	var p Project
	if err := FromJSON(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
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

func (pr *Proposal) ToJSON() (string, error) {
	return ToJSON(pr)
}
func ProposalFromJSON(data string) (*Proposal, error) {
	var pr Proposal
	if err := FromJSON(data, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

type VoteRecord struct {
	ProjectID   string `json:"project_id"`
	ProposalID  string `json:"proposal_id"`
	Voter       string `json:"voter"`
	ChoiceIndex []int  `json:"choice_index"` // indexes for options; for yes/no -> [0] or [1]
	Weight      int64  `json:"weight"`
	VotedAt     int64  `json:"voted_at"`
}

func (vr *VoteRecord) ToJSON() (string, error) {
	return ToJSON(vr)
}
func VoteRecordFromJSON(data string) (*VoteRecord, error) {
	var vr VoteRecord
	if err := FromJSON(data, &vr); err != nil {
		return nil, err
	}
	return &vr, nil
}
