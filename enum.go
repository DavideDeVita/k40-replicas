package main

// import "log"

// Define an enum with custom values
type Assurance float64

const (
	_2  Assurance = 0.99
	_2h Assurance = 0.995
	_3  Assurance = 0.999
	_3h Assurance = 0.9995
	_4  Assurance = 0.9999
	_4h Assurance = 0.99995
	_5  Assurance = 0.99999
	_5h Assurance = 0.999995
	_6  Assurance = 0.999999
	_6h Assurance = 0.9999995
	_7  Assurance = 0.9999999
)

var ASSURANCES []Assurance = []Assurance{_2, _2h, _3, _3h, _4, _4h, _5, _5h, _6, _6h, _7}

func rand_Assurance() Assurance{
	r := rand_ab_int(0, len(ASSURANCES))
	return ASSURANCES[r]
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
