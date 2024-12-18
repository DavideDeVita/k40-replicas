package main

import "fmt"

// Define an enum with custom values
type _Assurance float64

const (
	_2  _Assurance = 0.99
	_2h _Assurance = 0.995
	_3  _Assurance = 0.999
	_3h _Assurance = 0.9995
	_4  _Assurance = 0.9999
	_4h _Assurance = 0.99995
	_5  _Assurance = 0.99999
	_5h _Assurance = 0.999995
	_6  _Assurance = 0.999999
	_6h _Assurance = 0.9999995
	_7  _Assurance = 0.9999999
)

var _AssuranceS []_Assurance = []_Assurance{_2, _2h, _3, _3h, _4, _4h, _5, _5h, _6, _6h, _7}

func rand__Assurance(rt bool) _Assurance {
	var slice_start int = 0
	if rt {
		slice_start = 4
	}
	r := rand_ab_int(slice_start, len(_AssuranceS))
	return _AssuranceS[r]
}

func (a _Assurance) value() float64 {
	return float64(a)
}

func (a _Assurance) String() string {
	switch a {
	case _2:
		return fmt.Sprintf("2 nines (%.2f)", a)
	case _2h:
		return fmt.Sprintf("2 nines and a half (%.3f)", a)
	case _3:
		return fmt.Sprintf("3 nines (%.3f)", a)
	case _3h:
		return fmt.Sprintf("3 nines and a half (%.4f)", a)
	case _4:
		return fmt.Sprintf("4 nines (%.4f)", a)
	case _4h:
		return fmt.Sprintf("4 nines and a half (%.5f)", a)
	case _5:
		return fmt.Sprintf("5 nines (%.5f)", a)
	case _5h:
		return fmt.Sprintf("5 nines and a half (%.6f)", a)
	case _6:
		return fmt.Sprintf("6 nines (%.6f)", a)
	case _6h:
		return fmt.Sprintf("6 nines and a half (%.7f)", a)
	case _7:
		return fmt.Sprintf("7 nines (%.7f)", a)
	}
	return "Unknown"
}

// _Criticality is the lowest acceptable probability (threshold) that at least half succeed
//	 (for each solution, we compute the prob that at least half do not fail, if this prob is lower than this value, is rejected)
type _Criticality float64

const (
	NoCriticality      _Criticality = 0.95          //1.5
	LowCriticality     _Criticality = 0.999         //3
	MidLowCriticality  _Criticality = 0.999_95      //4.5
	MidCriticality     _Criticality = 0.999_999     //6
	MidHighCriticality _Criticality = 0.999_999_95  //7.5
	HighCriticality    _Criticality = 0.999_999_999 //9
)

func (c _Criticality) String() string {
	switch c {
	case NoCriticality:
		return fmt.Sprintf("No (%.2f)", c)
	case LowCriticality:
		return fmt.Sprintf("Low (%.3f)", c)
	case MidLowCriticality:
		return fmt.Sprintf("Mid/Low (%.5f)", c)
	case MidCriticality:
		return fmt.Sprintf("Mid (%.6f)", c)
	case MidHighCriticality:
		return fmt.Sprintf("Mid/High (%.8f)", c)
	case HighCriticality:
		return fmt.Sprintf("High (%.9f)", c)
	}
	return "Unknown"
}

func (c _Criticality) value() float64 {
	return float64(c)
}

// LogLevel
type LogLevel int

const (
	Log_None   LogLevel = 0
	Log_Scores LogLevel = 1
	Log_Some   LogLevel = 2
	Log_All    LogLevel = 3
)
