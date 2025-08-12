# Oinoko DAO – User Guide

This smart contract powers community-driven projects on the vsc network. It allows users to create projects, join them, make proposals, vote on proposals, and manage shared funds — all in a transparent, on-chain way powered by vsc smart contracts.

---

## 1. What You Can Do

- **Create a Project**  
  Start your own community project with a name, description, voting rules, and a shared treasury.
  You decide:
  - Who can make proposals (just you, or all members)
  - How members join (fixed fee for equal voting or stake-based for weighted voting)
  - The percentage of votes needed for approval
  - Proposal cost (goes into project funds)
  - Proposal duration
  - Minimum/Exact join amounts needed for users to join
  - Optional: Enable/disable features like reward distribution, secret voting, and more

- **Join a Project**  
  Become a member by sending the required join amount (set by the project).
  - In **Democratic voting** projects, every member’s vote counts equally.
  - In **Stake-based voting**, your vote weight depends on your contribution amount.

- **Make a Proposal**  
  Suggest an action or ask the community a question.
  Proposal types:
  - **Yes/No** (can also execute fund transfers if approved)
  - **Single Choice Poll**
  - **Multiple Choice Poll**
  
  Every proposal has:
  - Title, description, and extra metadata (for future features)
  - Duration for voting
  - Receiver (only for Yes/No fund transfers)
  - Cost (defined by the project, goes into treasury)

- **Vote on Proposals**  
  Members vote according to the project’s rules.  
  The system calculates results automatically after the deadline.

- **Send Additional Funds**  
  You can add more funds to the project’s treasury at any time to help the community achieve its goals.

---

## 2. How Voting Works

1. **Democratic Voting** – Every member has **1 vote**, regardless of stake.  
2. **Stake-based Voting** – Your vote weight = the amount you staked when joining.  

Projects can set:
- Minimum/Exact join amount
- Percentage needed to pass a proposal
- Who is allowed to create proposals
- If proposals require a cost

---

## 3. Proposal Results

When voting ends:
- If **Yes/No** and passes → Funds are sent (if applicable).  
- If a poll → Results are recorded on-chain for everyone to see.  
- The project treasury is updated automatically.

---

## 4. Commands & Actions

From your wallet or UI, you can:

| Action                 | Description |
|------------------------|-------------|
| `projects_create`      | Start a new project |
| `projects_get_all`     | View all projects |
| `projects_get_one`     | View one project |
| `projects_join`        | Become a member by paying the join amount |
| `projects_add_funds`   | Add fuds to the project's treasury 
| `proposals_create`     | Suggest an action/question |
| `proposals_get_all`    | View all proposals for a project |
| `proposals_get_one`    | View proposal |
| `proposals_vote`       | Cast your vote |
| `proposals_tally`      | Compute Pass / Fail / Quorum proposal state |
| `proposals_execute`    | Execute proposal if transfer or meta proposal |


---

## 5. Tips for Success

- Make your **proposal descriptions clear** so members know exactly what they are voting for.
- Always check the **voting deadline** before you submit your vote.
- If joining a stake-based project, your **initial stake matters** — it defines your voting power.
- Proposal costs go to the project treasury, so even failed proposals contribute to the community.

---

## 6. Safety

- All votes and results are public and stored on-chain.  
- The project owner can hand over control to another member via a special transfer function.  
- Only members can vote, and only according to the project’s rules.
