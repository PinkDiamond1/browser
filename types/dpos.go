package types

import "math/big"

type DposRegisterCandidate struct {
	URL string
}

type DposUpdateCandidate struct {
	URL string
}

type DposVoteCandidate struct {
	Candidate string
	Stake     *big.Int
}

type DposKickedCandidate struct {
	Candidates []string
}

type DposIrreversible struct {
	Reversible           uint64 `json:"reversible"`
	ProposedIrreversible uint64 `json:"proposedIrreversible"`
	BftIrreversible      uint64 `json:"bftIrreversible"`
}

type DposUpdateCandidatePubKey struct {
	PubKey PubKey
}
