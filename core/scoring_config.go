package core

import (
	"fmt"
	"strings"
)

// CONST STRINGS
const ASSURANCE_WASTELESS string = "assuranceWasteless"
const ENERGY_COST string = "energyCost"
const RESOURCES_FIT string = "resourcesFit"
const RT_WATELESS string = "rtWasteless"
const LIMIT_OVERCOMMIT_PENALTY string = "limitOvercommitPenalty"

// ScoringConfig definisce i parametri di scoring letti dal JSON.
type ScoringConfig struct {
	ResourceFit string             `json:"resourceFit"` // leastAllocated, mostAllocated, requestedToCapacityRatio
	Weights     map[string]float32 `json:"weights"`     // es: {"resourceFit":4, "energyCost":2, ...}
}

// normalizeWeights normalizza i pesi in modo che la somma sia 1. Metodo Private perchè lo usa solo Build...
func (c *ScoringConfig) normalizeWeights() {
	if c.Weights == nil {
		c.Weights = make(map[string]float32)
	}

	c.Weights = normalizeAlgorithmKeys(c.Weights)

	var sum float32
	for _, v := range c.Weights {
		sum += v
	}
	if sum == 0 {
		c.Weights[RESOURCES_FIT] = 1
		sum = 1
	}
	for k, v := range c.Weights {
		c.Weights[k] = v / sum
	}
}

func normalizeAlgorithmKeys(weights map[string]float32) map[string]float32 {
	var aliases map[string]string = map[string]string{
		"assurancewasteless": ASSURANCE_WASTELESS,
		"assurance":          ASSURANCE_WASTELESS,

		"energycost":    ENERGY_COST,
		"energy":        ENERGY_COST,
		"energetic":     ENERGY_COST,
		"energeticcost": ENERGY_COST,

		"resourcefit":  RESOURCES_FIT,
		"resourcesfit": RESOURCES_FIT,
		"resources":    RESOURCES_FIT,
		"resource":     RESOURCES_FIT,
		"k8s":          RESOURCES_FIT,
		"kubernetes":   RESOURCES_FIT,

		"rtwasteless": RT_WATELESS,
		"rt":          RT_WATELESS,
		"realtime":    RT_WATELESS,

		"limitovercommitpenalty": LIMIT_OVERCOMMIT_PENALTY,
		"overcommitpenalty":      LIMIT_OVERCOMMIT_PENALTY,
		"limitovercommit":        LIMIT_OVERCOMMIT_PENALTY,
		"limitpenalty":           LIMIT_OVERCOMMIT_PENALTY,
	}

	norm := make(map[string]float32)
	for k, v := range weights {
		lower := strings.ToLower(k)
		lower = strings.ReplaceAll(lower, "_", "")
		lower = strings.ReplaceAll(lower, "-", "")
		// Verifica se esiste un alias per questa chiave normalizzata
		if alias, ok := aliases[lower]; ok {
			norm[alias] = v
		} else {
			// opzionale: loggare o ignorare in silenzio
			fmt.Printf("[WARNING] Nome di funzione sconosciuto nella dichiarazione dei pesi: %s\tIgnorato\n", k)
		}
	}
	return norm
}

// BuildScoringFunction costruisce la funzione di scoring normalizzata.
func BuildScoringFunction(cfg ScoringConfig,
	hp AlgorithmHyperparams,
) (func(*WorkerNode, *Pod) float32, func(n *WorkerNode, p *Pod) map[string]float32) {
	cfg.normalizeWeights()

	// Seleziona la funzione di ResourceFit
	var fitFunc func(*WorkerNode, *Pod) float32
	switch strings.ToLower(cfg.ResourceFit) {
	case "mostallocated":
		fitFunc = K8sMostAllocatedScore
	case "requestedtocapacityratio":
		fitFunc = K8sRequestedToCapacityRatioScore
	case "leastallocated":
		fitFunc = K8sLeastAllocatedScore
	default: // leastAllocated di default
		fitFunc = K8sLeastAllocatedScore
		fmt.Printf("[WARN] Unknown ResourceFit '%s'. Using 'LeastAllocated\n", cfg.ResourceFit)
	}

	scoringFunc := func(n *WorkerNode, p *Pod) float32 {
		score := float32(0)

		if w, ok := cfg.Weights[RESOURCES_FIT]; ok && w > 0 {
			score += w * fitFunc(n, p)
		}
		if w, ok := cfg.Weights[ENERGY_COST]; ok && w > 0 {
			switch hp.EnergyCostMode {
			case "delta", "deltaabsolute", "absolutedelta":
				score += w * DeltaEnergyCostRatio_absolute(n, p)
			case "linear":
				score += w * EnergyCostRatio_linear(n, p)
			case "lineardelta", "deltalinear":
				score += w * DeltaEnergyCostRatio_linear(n, p)
			default: // absolute
				score += w * EnergyCostRatio_absolute(n, p)
			}
		}
		if w, ok := cfg.Weights[ASSURANCE_WASTELESS]; ok && w > 0 {
			score += w * SigmaAssuranceWasteless(n, p)
		}
		if w, ok := cfg.Weights[RT_WATELESS]; ok && w > 0 {
			score += w * RTWaste(n, p)
		}
		if w, ok := cfg.Weights[LIMIT_OVERCOMMIT_PENALTY]; ok && w > 0 {
			score += w * LimitOvercommitPenalty(n, p)
		}

		return score
	}

	explainScoringFunc := func(n *WorkerNode, p *Pod) map[string]float32 {
		var subscores map[string]float32 = map[string]float32{}

		if w, ok := cfg.Weights[RESOURCES_FIT]; ok && w > 0 {
			subscores[RESOURCES_FIT] = w * fitFunc(n, p)
		}
		if w, ok := cfg.Weights[ENERGY_COST]; ok && w > 0 {
			switch hp.EnergyCostMode {
			case "delta", "deltaabsolute", "absolutedelta":
				subscores[ENERGY_COST] = w * DeltaEnergyCostRatio_absolute(n, p)
			case "linear":
				subscores[ENERGY_COST] = w * EnergyCostRatio_linear(n, p)
			case "lineardelta", "deltalinear":
				subscores[ENERGY_COST] = w * DeltaEnergyCostRatio_linear(n, p)
			default: // absolute
				subscores[ENERGY_COST] = w * EnergyCostRatio_absolute(n, p)
			}
		}
		if w, ok := cfg.Weights[ASSURANCE_WASTELESS]; ok && w > 0 {
			subscores[ASSURANCE_WASTELESS] = w * SigmaAssuranceWasteless(n, p)
		}
		if w, ok := cfg.Weights[RT_WATELESS]; ok && w > 0 {
			subscores[RT_WATELESS] = w * RTWaste(n, p)
		}
		if w, ok := cfg.Weights[LIMIT_OVERCOMMIT_PENALTY]; ok && w > 0 {
			subscores[LIMIT_OVERCOMMIT_PENALTY] = w * LimitOvercommitPenalty(n, p)
		}

		return subscores
	}

	return scoringFunc, explainScoringFunc
}
