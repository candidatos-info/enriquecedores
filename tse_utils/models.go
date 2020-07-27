package tseutils

// Candidate is the struct for a candidate on db
type Candidate struct {
	Email        string
	State        string
	City         string
	Name         string
	Party        string
	BallotNumber int
	LegalCode    string
	Password     string
}

// CandidatureData is the data related to one election
type CandidatureData struct {
	CandidateLegalCode string
	SquencialCandidate string
	SiteURL            string
	Facebook           string
	Twitter            string
	Year               int
	Proposal           string
}
