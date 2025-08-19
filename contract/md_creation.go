package contract

import (
	"strconv"
	"strings"
	"time"
)

// DescribeProposalMarkdown returns a markdown-formatted summary of a proposal,
// including vote counts, participation, and individual votes in table format.
//
//go:wasmexport proposal_describe
func DescribeProposalMarkdown(proposalID string) *string {
	state := getState()
	prpsl, err := loadProposal(state, proposalID)
	if err != nil {
		return returnJsonResponse(
			"proposal_describe", false, map[string]interface{}{
				"details": "propsal not found",
			},
		)
	}

	md := "### Proposal: " + prpsl.Name + "\n\n"
	md += "| Field | Value |\n"
	md += "|-------|-------|\n"
	md += "| ID | " + prpsl.ID + " |\n"
	md += "| Description | " + prpsl.Description + " |\n"
	if prpsl.JsonMetadata != "" {
		md += "| Metadata | " + prpsl.JsonMetadata + " |\n"
	}
	md += "| Type | " + string(prpsl.Type) + " |\n"
	md += "| State | " + string(prpsl.State) + " |\n"
	md += "| Passed | " + strconv.FormatBool(prpsl.Passed) + " |\n"
	md += "| Created At | " + time.Unix(prpsl.CreatedAt, 0).Format(time.RFC3339) + " |\n"
	if prpsl.FinalizedAt > 0 {
		md += "| Finalized At | " + time.Unix(prpsl.FinalizedAt, 0).Format(time.RFC3339) + " |\n"
	}
	if prpsl.PassTimestamp > 0 {
		md += "| Pass Timestamp | " + time.Unix(prpsl.PassTimestamp, 0).Format(time.RFC3339) + " |\n"
	}
	if prpsl.Receiver != "" {
		md += "| Receiver | " + prpsl.Receiver + " |\n"
	}

	// Load votes
	votes := loadVotesForProposal(state, prpsl.ProjectID, proposalID)
	totalParticipation := int64(0)
	optionCounts := make(map[int]int64)
	for _, v := range votes {
		totalParticipation += v.Weight
		for _, idx := range v.ChoiceIndex {
			optionCounts[idx] += v.Weight
		}
	}

	md += "| Total Participation (weight) | " + strconv.FormatInt(totalParticipation, 10) + " |\n"

	if len(optionCounts) > 0 {
		md += "\n**Option Counts:**\n\n"
		md += "| Option Index | Votes (weight) |\n"
		md += "|-------------|----------------|\n"
		for idx, count := range optionCounts {
			md += "| " + strconv.Itoa(idx) + " | " + strconv.FormatInt(count, 10) + " |\n"
		}
		md += "\n"
	}

	if len(votes) > 0 {
		md += "**Individual Votes:**\n\n"
		md += "| Voter | Weight | Choice Index(es) |\n"
		md += "|-------|--------|----------------|\n"
		for _, v := range votes {

			choices := make([]string, len(v.ChoiceIndex))
			for i, c := range v.ChoiceIndex {
				choices[i] = strconv.Itoa(c)
			}
			md += "| " + v.Voter + " | " + strconv.FormatInt(v.Weight, 10) + " | " + strings.Join(choices, ", ") + " |\n"
		}
		md += "\n"
	}

	return returnJsonResponse(
		"proposal_describe", true, map[string]interface{}{
			"details": md,
		},
	)
}

// DescribeProjectMarkdown returns a markdown-formatted summary of a project.
// If includeProposals is true, it will embed all proposals for this project in table format.
//
//go:wasmexport project_describe
func DescribeProjectMarkdown(projectID string, includeProposals bool) *string {
	state := getState()
	prj, err := loadProject(state, projectID)
	if err != nil {
		return returnJsonResponse(
			"project_describe", false, map[string]interface{}{
				"details": "project not found",
			},
		)
	}

	md := "# Project: " + prj.Name + "\n\n"
	md += "| Field | Value |\n"
	md += "|-------|-------|\n"
	md += "| ID | " + prj.ID + " |\n"
	md += "| Description | " + prj.Description + " |\n"
	if prj.JsonMetadata != "" {
		md += "| Metadata | " + prj.JsonMetadata + " |\n"
	}
	md += "| Owner | " + prj.Owner + " |\n"
	md += "| Created At | " + time.Unix(prj.CreatedAt, 0).Format(time.RFC3339) + " |\n"
	md += "| Voting System | " + string(prj.Config.VotingSystem) + " |\n"
	md += "| Quorum Percent | " + strconv.Itoa(prj.Config.QuorumPercent) + "% |\n"
	md += "| Threshold Percent | " + strconv.Itoa(prj.Config.ThresholdPercent) + "% |\n"
	md += "| Proposal Cost | " + strconv.FormatInt(prj.Config.ProposalCost, 10) + " |\n"
	md += "| Proposal Duration | " + strconv.FormatInt(prj.Config.ProposalDurationSecs, 10) + " seconds |\n"
	md += "| Project Funds | " + strconv.FormatInt(prj.Funds, 10) + " |\n"

	if includeProposals {
		proposalIds := listProposalIDsForProject(state, projectID)

		if len(proposalIds) > 0 {
			md += "\n## Proposals\n\n"
			for _, pid := range proposalIds {
				propResult := DescribeProposalMarkdown(pid)

				if propResult == nil {
					md += "*Error: proposal " + pid + " returned nil*\n\n"
					continue
				}

				var propJson map[string]interface{}
				if err := FromJSON(*propResult, &propJson); err != nil {
					md += "*Error loading proposal " + pid + ": " + err.Error() + "*\n\n"
					continue
				}

				if details, ok := propJson["details"].(string); ok {
					md += details + "\n\n"
				} else {
					md += "*Error: proposal " + pid + " missing details*\n\n"
				}
			}
		}
	}

	return returnJsonResponse(
		"project_describe", true, map[string]interface{}{
			"details": md,
		},
	)
}
