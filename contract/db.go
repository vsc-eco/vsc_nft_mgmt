package main

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// Contract State Persistence helpers
////////////////////////////////////////////////////////////////////////////////

func saveProject(state State, pro *Project) {
	key := projectKey(pro.ID)
	b, _ := json.Marshal(pro)
	state.Set(key, string(b))
}

func loadProject(state State, id string) (*Project, error) {
	key := projectKey(id)
	ptr := state.Get(key)
	if ptr == nil {
		return nil, fmt.Errorf("project %s not found", id)
	}
	var pro Project
	if err := json.Unmarshal([]byte(*ptr), &pro); err != nil {
		return nil, fmt.Errorf("failed unmarshal project %s: %v", id, err)
	}
	return &pro, nil
}

func addProjectToIndex(state State, id string) {
	ptr := state.Get(projectsIndexKey)
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
	state.Set(projectsIndexKey, string(b))
}

func listAllProjectIDs(state State) []string {
	ptr := state.Get(projectsIndexKey)
	if ptr == nil {
		return []string{}
	}
	var ids []string
	if err := json.Unmarshal([]byte(*ptr), &ids); err != nil {
		return []string{}
	}
	return ids
}

func saveProposal(state State, prpsl *Proposal) {
	key := proposalKey(prpsl.ID)
	b, _ := json.Marshal(prpsl)
	state.Set(key, string(b))
}

func loadProposal(state State, id string) (*Proposal, error) {
	key := proposalKey(id)
	ptr := state.Get(key)
	if ptr == nil {
		return nil, fmt.Errorf("proposal %s not found", id)
	}
	var prpsl Proposal
	if err := json.Unmarshal([]byte(*ptr), &prpsl); err != nil {
		return nil, fmt.Errorf("failed unmarshal proposal %s: %v", id, err)
	}
	return &prpsl, nil
}

func addProposalToProjectIndex(state State, projectID, proposalID string) {
	key := projectProposalsIndexKey(projectID)
	ptr := state.Get(key)
	var ids []string
	if ptr != nil {
		json.Unmarshal([]byte(*ptr), &ids)
	}
	// TODO: avoid duplicates
	ids = append(ids, proposalID)
	b, _ := json.Marshal(ids)
	state.Set(key, string(b))
}

func listProposalIDsForProject(state State, projectID string) []string {
	key := projectProposalsIndexKey(projectID)
	ptr := state.Get(key)
	if ptr == nil {
		return []string{}
	}
	var ids []string
	if err := json.Unmarshal([]byte(*ptr), &ids); err != nil {
		return []string{}
	}
	return ids
}

func saveVote(state State, vote *VoteRecord) {
	key := voteKey(vote.ProjectID, vote.ProposalID, vote.Voter)
	b, _ := json.Marshal(vote)
	state.Set(key, string(b))

	// ensure voter listed in index for iteration (store list under project:proposal:voters)
	votersKey := fmt.Sprintf("proposal:%s:%s:voters", vote.ProjectID, vote.ProposalID)
	ptr := state.Get(votersKey)
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
		state.Set(votersKey, string(nb))
	}
}

func loadVotesForProposal(state State, projectID, proposalID string) []VoteRecord {
	votersKey := fmt.Sprintf("proposal:%s:%s:voters", projectID, proposalID)
	ptr := state.Get(votersKey)
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
		vp := state.Get(vk)
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
func removeVote(state State, projectID, proposalID, voter string) {
	key := voteKey(projectID, proposalID, voter)
	state.Delete(key)

	// remove from voter list
	votersKey := fmt.Sprintf("proposal:%s:%s:voters", projectID, proposalID)
	ptr := state.Get(votersKey)
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
	state.Set(votersKey, string(nb))
}
