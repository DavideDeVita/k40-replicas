package main

// import "log"

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

// LogLevel
type LogLevel int

const (
	Log_None   LogLevel = 0
	Log_Scores LogLevel = 1
	Log_Some   LogLevel = 2
	Log_All    LogLevel = 3
)

// Interference
const _LIGHT_INTERFERENCE_CHANCE float32 = 0.1
const _HEAVY_INTERFERENCE_CHANCE float32 = 0.01

type Interference int

const (
	No_Interference    Interference = 0
	Light_Interference Interference = 1
	Heavy_Interference Interference = 2
)
