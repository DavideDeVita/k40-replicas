package main

import (
	"fmt"
	"strconv"
	"strings"
)

/*	 ASSURANCE	 */
// var _AssuranceS []float32 = []float32{0.99, 0.995, 0.999, 0.9995, 0.9999, 0.99995, 0.99999, 0.999995, 0.999999, 0.9999995, 0.9999999}
var _AssuranceValues []float32 = []float32{
	0.1, 0.25, 0.33, 0.5, // Solo non rt	(!rt: 5 5 5 5)			Tot prob !rt: 31
	0.66, 0.75, 0.8, 0.85, //	(!rt: 4, 3, 2, 2)	(rt: 3, 4, 4, 3)
	0.9, 0.925, 0.95, 0.975, 0.99, // Solo rt				(rt: 2, 1, 1, 1, 1)		 Tot prob rt: 20
}

func random_Assurance(rt bool) float32 {
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
type _Criticality float32

const ( //									 Prob nonRt   (sum = 27)			prob Rt (22)
	NoCriticality      _Criticality = 0.   // 4
	MinCriticality     _Criticality = 0.5  // 5
	BarelyCriticality  _Criticality = 0.66 // 5
	LowCriticality     _Criticality = 0.75 // 4								3
	MidLowCriticality  _Criticality = 0.8  // 4								4
	MidCriticality     _Criticality = 0.9  // 3								4
	MidHighCriticality _Criticality = 0.95 // 2								4
	HighCriticality    _Criticality = 0.99 // 									3
	// ExtraCriticality   _Criticality = 0.995 // 									2
	MaxCriticality _Criticality = 0.995 // 									2
)

var _Criticality_IdxLookup map[_Criticality]int = map[_Criticality]int{
	NoCriticality:      0,
	MinCriticality:     1,
	BarelyCriticality:  2,
	LowCriticality:     3,
	MidLowCriticality:  4,
	MidCriticality:     5,
	MidHighCriticality: 6,
	HighCriticality:    7,
	// ExtraCriticality:   8,
	MaxCriticality: 8,
}

func random_Criticality(rt bool) _Criticality {
	rnd := rand_01()
	if rt {
		if rnd < 2./20. {
			return MaxCriticality
			// } else if rnd < 4./22. { //+2
			// 	return ExtraCriticality
		} else if rnd < 5./20. { //+3
			return HighCriticality
		} else if rnd < 9./20. { //+4
			return MidHighCriticality
		} else if rnd < 13./20. { //+4
			return MidCriticality
		} else if rnd < 17./20. { //+4
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
			return MinCriticality
		} else { //+4
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
	// case ExtraCriticality:
	// 	return fmt.Sprintf("Extra (%.3f)", c)
	case MaxCriticality:
		return fmt.Sprintf("Max (%.3f)", c)
	}
	return "Unknown"
}

func (c _Criticality) value() float32 {
	return float32(c)
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
func __make_InterferenceCounter(numTests int, numPods int) ([][]float32, [][]string) {
	var ret_f [][]float32 = make([][]float32, len(_Criticality_IdxLookup))
	var ret_s [][]string = make([][]string, numPods)
	for c, i := range _Criticality_IdxLookup {
		ret_f[i] = make([]float32, numTests+1)
		ret_f[i][0] = c.value()
		for j := 1; j <= numTests; j++ {
			ret_f[i][j] = 0.
		}
	}

	for i := 0; i < numPods; i++ {
		ret_s[i] = make([]string, numTests+2)
		ret_s[i][0] = fmt.Sprintf("%d", i)
		for j := 1; j < numTests+2; j++ {
			ret_s[i][j] = ""
		}
	}

	return ret_f, ret_s
}

func recordInterference(pid int, algo_idx int, wnid int) {
	var algostr string = Intereferences_byC_byWN[pid][algo_idx+2]
	parts := strings.Split(algostr, "#")
	if len(parts) != 3 {
		panic("invalid format")
	}

	ids := strings.Split(parts[1], ",")
	counts := strings.Split(parts[2], "::")
	tot := counts[1]
	counts = strings.Split(counts[0], ",")
	// log.Printf("Record interference: pid %d, algo %d, wnid %d, str: %s\n", pid, algo_idx, wnid, algostr)
	wnidStr := fmt.Sprintf("%d", wnid)
	pos := -1
	for i, id := range ids {
		if id == wnidStr {
			pos = i
			break
		}
	}

	if pos != -1 {
		val, err := strconv.Atoi(counts[pos])
		if err != nil {
			panic("invalid interference count")
		}
		val += 1
		counts[pos] = strconv.Itoa(val)

		// Tot
		val, err = strconv.Atoi(tot)
		if err != nil {
			panic("invalid interference tot count")
		}
		val += 1
		tot = strconv.Itoa(val)
	}

	// Reconstruct the string
	newAlgostr := fmt.Sprintf("%s#%s#%s::%s", parts[0], strings.Join(ids, ","), strings.Join(counts, ","), tot)
	Intereferences_byC_byWN[pid][algo_idx+2] = newAlgostr
}

func recordDeployment(pid int, algo_idx int, sol Solution) {
	if !sol.rejected {
		var wnids = sol.list_Ids()
		var s string = fmt.Sprintf("%d#", len(wnids))
		var fmtids = ""
		var fmtcounter = ""
		for _, wnid := range wnids {
			fmtids += fmt.Sprintf("%d,", wnid)
			fmtcounter += "0,"
		}
		s += fmtids[:len(fmtids)-1] + "#" + fmtcounter[:len(fmtcounter)-1] + "::0"
		Intereferences_byC_byWN[pid][algo_idx+2] = s
	}
}
