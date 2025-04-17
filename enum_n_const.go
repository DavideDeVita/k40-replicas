package main

import "fmt"

/*	 ASSURANCE	 */
// var _AssuranceS []float64 = []float64{0.99, 0.995, 0.999, 0.9995, 0.9999, 0.99995, 0.99999, 0.999995, 0.999999, 0.9999995, 0.9999999}
var _AssuranceValues []float64 = []float64{
	0.1, 0.25, 0.33, 0.5, // Solo non rt	(!rt: 5 5 5 5)			Tot prob !rt: 31
	0.66, 0.75, 0.8, 0.85, //	(!rt: 4, 3, 2, 2)	(rt: 3, 4, 4, 3)
	0.9, 0.925, 0.95, 0.975, 0.99, // Solo rt				(rt: 2, 1, 1, 1, 1)		 Tot prob rt: 20
}

func random_Assurance(rt bool) float64 {
	rnd := rand_01()
	if !rt {
		if rnd < 2./31. { //+2
			return 0.85
		} else if rnd < 4./31. { //+2
			return 0.8
		} else if rnd < 7./31. { //+3
			return 0.75
		} else if rnd < 11./31. { //+4
			return 0.66
		} else if rnd < 16./31. { //+5
			return 0.5
		} else if rnd < 21./31. { //+5
			return 0.33
		} else if rnd < 26./31. { //+5
			return 0.25
		} else { //+5
			return 0.1
		}
	} else {
		if rnd < 1./20. { //+1
			return 0.99
		} else if rnd < 2./20. { //+1
			return 0.975
		} else if rnd < 3./20. { //+1
			return 0.95
		} else if rnd < 4./20. { //+1
			return 0.925
		} else if rnd < 6./20. { //+2
			return 0.9
		} else if rnd < 9./20. { //+3
			return 0.85
		} else if rnd < 13./20. { //+4
			return 0.8
		} else if rnd < 17./20. { //+4
			return 0.75
		} else { //+3
			return 0.66
		}
	}
}

/*	 CRITICALITY	 */
// _Criticality is the lowest acceptable probability (threshold) that at least half succeed
//	 (for each solution, we compute the prob that at least half do not fail, if this prob is lower than this value, is rejected)
type _Criticality float64

const ( //									 Prob nonRt   (sum = 27)			prob Rt (22)
	NoCriticality      _Criticality = 0.    // 4
	MinCriticality     _Criticality = 0.5   // 5
	BarelyCriticality  _Criticality = 0.66  // 5
	LowCriticality     _Criticality = 0.75  // 4								3
	MidLowCriticality  _Criticality = 0.8   // 4								4
	MidCriticality     _Criticality = 0.9   // 3								4
	MidHighCriticality _Criticality = 0.95  // 2								4
	HighCriticality    _Criticality = 0.99  // 									3
	ExtraCriticality   _Criticality = 0.995 // 									2
	MaxCriticality     _Criticality = 0.999 // 									2
)

var _CriticalityValues []_Criticality = []_Criticality{
	NoCriticality, MinCriticality, BarelyCriticality,
	LowCriticality, MidLowCriticality, MidCriticality, MidHighCriticality, HighCriticality,
	ExtraCriticality, MaxCriticality,
}

func random_Criticality(rt bool) _Criticality {
	rnd := rand_01()
	if rt {
		if rnd < 2./22. {
			return MaxCriticality
		} else if rnd < 4./22. { //+2
			return ExtraCriticality
		} else if rnd < 7./22. { //+3
			return HighCriticality
		} else if rnd < 11./22. { //+4
			return MidHighCriticality
		} else if rnd < 15./22. { //+4
			return MidCriticality
		} else if rnd < 19./22. { //+4
			return MidLowCriticality
		} else { //+3
			return LowCriticality
		}
	} else {
		if rnd < 2./27. {
			return MidHighCriticality
		} else if rnd < 5./27. { //+3
			return MidCriticality
		} else if rnd < 9./27. { //+4
			return MidLowCriticality
		} else if rnd < 13./27. { //+4
			return LowCriticality
		} else if rnd < 18./27. { //+5
			return BarelyCriticality
		} else if rnd < 23./27. { //+5
			return LowCriticality
		} else { //+5
			return NoCriticality
		}
	}
}

func (c _Criticality) String() string {
	switch c {
	case NoCriticality:
		return fmt.Sprintf("No (%.1f)", c)
	case MinCriticality:
		return fmt.Sprintf("Min (%.1f)", c)
	case BarelyCriticality:
		return fmt.Sprintf("Barely (%.2f)", c)
	case LowCriticality:
		return fmt.Sprintf("Low (%.2f)", c)
	case MidLowCriticality:
		return fmt.Sprintf("Mid/Low (%.2f)", c)
	case MidCriticality:
		return fmt.Sprintf("Mid (%.2f)", c)
	case MidHighCriticality:
		return fmt.Sprintf("Mid/High (%.3f)", c)
	case HighCriticality:
		return fmt.Sprintf("High (%.3f)", c)
	case ExtraCriticality:
		return fmt.Sprintf("Extra (%.3f)", c)
	case MaxCriticality:
		return fmt.Sprintf("High (%.3f)", c)
	}
	return "Unknown"
}

func (c _Criticality) value() float64 {
	return float64(c)
}

/*	 Log Level	 */
type LogLevel int

const (
	Log_None   LogLevel = 0
	Log_Some   LogLevel = 1
	Log_Scores LogLevel = 2
	Log_All    LogLevel = 3
)

/*	 Interference Map	 */
func __emptyIMap() map[float64]int {
	var ret map[float64]int = map[float64]int{}
	for _, c := range _CriticalityValues {
		ret[c.value()] = 0
	}
	return ret
}
