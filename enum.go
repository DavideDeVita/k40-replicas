package main

// import "fmt"

// Define an enum with custom values
type Assurance int

const (
	LowAssurance  Assurance = 1
	HighAssurance Assurance = 2
)

var ASSURANCES []Assurance = []Assurance{LowAssurance, HighAssurance}

func (a Assurance) String() string {
	if a == LowAssurance {
		return "Low"
	} else {
		return "High"
	}
}

// Define an enum with custom values
type Criticality int

const (
	LowCriticality  Criticality = 1
	MidCriticality  Criticality = 2
	HighCriticality Criticality = 3
)

func (c Criticality) String() string {
	if c == LowCriticality {
		return "Low"
	} else if c == MidCriticality {
		return "Mid"
	} else {
		return "High"
	}
}


// Log
type Log int

const (
	Log_None  Log = 0
	Log_Scores Log = 1
	Log_Some  Log = 2
	Log_All Log = 3
)