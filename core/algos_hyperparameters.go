package core

import (
	"fmt"
	"strings"

	cmn "local/KOrche/common"
)

// AlgorithmHyperparams definisce i parametri di configurazione degli algoritmi di placement.
type AlgorithmHyperparams struct {
    MaxReplicas          			int     `json:"maxReplicas,omitempty"`
    DP_MaxNeighbors         		int     `json:"dp_maxNeighbors,omitempty"`
    DP_SolutionOversizeSearch     	int     `json:"dp_solutionOversizeSearch,omitempty"`
    DP_NeighborSpan         		int     `json:"dp_neighborSpan,omitempty"`
    DP_EnergyCostWakeupMultiplier 	float32 `json:"dp_energyCostWakeupMultiplier,omitempty"`
    DP_EnergyCostWakeupConst	 	float32 `json:"dp_energyCostWakeupConst,omitempty"`
    DP_ScoreAggregationMode     	string  `json:"dp_scoreAggregationMode,omitempty"`
	OvercommitResourceDangerRatio	float32 `json:"overcommit_ReourceDangerRatio,omitempty"`
	EnergyCostMode 					string  `json:"energyCostMode,omitempty"` // "absolute", "delta", "linear", "deltaLinear"
	Verbose							bool    `json:"verbose,omitempty"`
}

func DefaultAlgorithmHyperparams() AlgorithmHyperparams {
    return AlgorithmHyperparams{
        MaxReplicas:          			-1,
        DP_MaxNeighbors:         		10,
        DP_SolutionOversizeSearch:     	2,
        DP_NeighborSpan:         		2,
        DP_EnergyCostWakeupMultiplier: 	1.5,
        DP_EnergyCostWakeupConst: 		0.1,
        DP_ScoreAggregationMode:     	"SquaredSum",
		OvercommitResourceDangerRatio:	0.95,
		EnergyCostMode:					"absolute",
		Verbose:						false,
    }
}

// SetAlgorithmParams crea un AlgorithmParams a partire da una mappa generica (ad es. decodificata da JSON).
// I campi non presenti nella mappa vengono mantenuti ai valori di default.
func SetAlgorithmParams(raw map[string]interface{}) AlgorithmHyperparams {
	params := DefaultAlgorithmHyperparams()

	if raw == nil {
		return params
	}

	for key, value := range raw {
		fixkey := strings.ReplaceAll(strings.ToLower(key), "_", "")
		fixkey = strings.ReplaceAll(fixkey, "-", "")
		switch fixkey {
		case "maxreplicas":
			if v, ok := cmn.ToInt(value); ok {
				params.MaxReplicas = v
			}
		case "dpmaxneighbors":
			if v, ok := cmn.ToInt(value); ok {
				params.DP_MaxNeighbors = v
			}
		case "dpsolutionoversizesearch", "dpoverkillreplicas":
			if v, ok := cmn.ToInt(value); ok {
				params.DP_SolutionOversizeSearch = v
			}
		case "dpneighborspan":
			if v, ok := cmn.ToInt(value); ok {
				params.DP_NeighborSpan = v
			}
		case "dpenergycostwakeupmultiplier":
			if v, ok := cmn.ToFloat(value); ok {
				if v>=1{
					params.DP_EnergyCostWakeupMultiplier = v
				}else{
					fmt.Printf("[WARN] Ignored DP_EnergyCostWakeupMultiplier: %f.\n\tMust be >= 1\n", v)
				}
			}
		case "dpenergycostwakeupconst":
			if v, ok := cmn.ToFloat(value); ok {
				if v>=0 && v<1{
					params.DP_EnergyCostWakeupConst = v
				}else{
					fmt.Printf("[WARN] Ignored DP_EnergyCostWakeupConst: %f.\n\tMust be between 0 (incl) and 1 (excl)\n", v)
				}
			}
		case "dpscoreaggregationmode":
			if v, ok := value.(string); ok {
				params.DP_ScoreAggregationMode = v
			}
		case "overcommitreourcedangerratio":
			if v, ok := cmn.ToFloat(value); ok {
				params.OvercommitResourceDangerRatio = v
			}
		case "energycostmode":
			if v, ok := value.(string); ok {
				fixModeStr := strings.ReplaceAll(strings.ToLower(v), "_", "")
				fixModeStr = strings.ReplaceAll(fixModeStr, "-", "")
				params.EnergyCostMode = fixModeStr
			}
		case "verbose":
			if v, ok := value.(bool); ok {
				params.Verbose = v
			}
		default:
			fmt.Printf("[WARN] Ignored unknown algorithm param: %s : %v\n", key, value)
		}
	}

	return params
}