package core

import (
	"fmt"

	cmn "local/KOrche/common"
)

type Solution struct {
	Accepted    bool
	Pod         *Pod
	Replicas    int
	Probability float32
	Nodes       []PlacedNode
	DeltaEnergy float32
	Explanation string
}

type PlacedNode struct {
	_autoID       int
	NodeID        string
	CPU           ResourceStats
	Memory        ResourceStats
	Storage       ResourceStats
	Score         float32
	ExplainScores map[string]float32 // e.g. {"resource_fit": 0.7, "energy_cost": 0.2, ...}
}

func InitSolution(pod *Pod) Solution {
	return Solution{
		Accepted:    true, // optimistic default
		Pod:         pod,
		Replicas:    0,
		Probability: 0.0,
		Nodes:       []PlacedNode{},
		Explanation: "",
	}
}

func (s *Solution) Reject(reason string) {
	s.Accepted = false
	s.Nodes = nil
	s.Replicas = 0
	s.Probability = 0
	s.Explanation = fmt.Sprintf("Job %s rejected: %s", s.Pod.ID, reason)
}

func (s *Solution) AddToSolution(node *WorkerNode, update_prob float32, wn_score float32, explainScore map[string]float32) {
	s.Replicas++
	s.Probability = update_prob

	s.Nodes = append(s.Nodes, PlacedNode{
		_autoID: node.AutoID,
		NodeID:  node.ID,
		CPU: ResourceStats{
			Capacity:  node.CPU.Capacity,
			Requested: node.CPU.Requested + s.Pod.CPU.Requested,
			Limit:     node.CPU.Limit + s.Pod.CPU.Limit,
		},
		Memory: ResourceStats{
			Capacity:  node.Memory.Capacity,
			Requested: node.Memory.Requested + s.Pod.Memory.Requested,
			Limit:     node.Memory.Limit + s.Pod.Memory.Limit,
		},
		Storage: ResourceStats{
			Capacity:  node.Storage.Capacity,
			Requested: node.Storage.Requested + s.Pod.Storage.Requested,
			Limit:     node.Storage.Limit + s.Pod.Storage.Limit,
		},
		Score:         wn_score,
		ExplainScores: cmn.CloneMap(explainScore),
	})

}

// ComputeEnergyDelta calcola l'aumento del costo energetico
// dovuto all'attivazione di nodi Idle per questa soluzione.
func (s *Solution) ComputeEnergyDelta(cluster *Cluster, hp AlgorithmHyperparams) float32 {
	var delta float32 = 0.
	for _, placed := range s.Nodes {
		wn, state := cluster.GetNodeByID(placed._autoID)

		switch hp.EnergyCostMode {
		case "linear", "lineardelta", "deltalinear":
			delta += wn.ComputeLinearEnergyCost(s.Pod) - wn.ComputeLinearEnergyCost(nil)
		default: // "absolute", "delta", "deltaabsolute", "absolutedelta"
			if wn != nil && state == Idle {
				delta += float32(wn.EnergyCostCoef - wn.EnergyCostIdle)
			}
		}
	}
	return delta
}

func (s *Solution) WrapUpSolution(cluster *Cluster, hyperparams AlgorithmHyperparams) {
	if !s.Accepted {
		return
	}

	s.DeltaEnergy = s.ComputeEnergyDelta(cluster, hyperparams)

	replicas_s := "replica"
	if s.Replicas != 1 {
		replicas_s += "s"
	}

	// Removed
	if hyperparams.Verbose {
		var new_EnergyCost float32 = CurrClusterEnergyCost + s.DeltaEnergy
		deltaEnergy_S := fmt.Sprintf("Cluster energy cost (%s)", hyperparams.EnergyCostMode)
		if s.DeltaEnergy == 0 {
			deltaEnergy_S += fmt.Sprintf(" unchanged: %.2f", CurrClusterEnergyCost)
		} else {
			deltaEnergy_S += fmt.Sprintf(" changed from %.2f to %.2f", CurrClusterEnergyCost, new_EnergyCost)
		}

		prob_i := s.Probability * 100.
		crit_i := s.Pod.Criticality * 100.

		s.Explanation = fmt.Sprintf("Job %s accepted:\n"+
			"\tCriticality constraint met (%.2f%% / %.2f%%)\n"+
			"\tSolution found with %d %s\n"+
			"\t%s\n",
			s.Pod.ID, prob_i, crit_i, s.Replicas, replicas_s, deltaEnergy_S)
	} else{
		// var crit_i int = int(s.Pod.Criticality*100.)
		var crit_100 float32 = s.Pod.Criticality*100.
		s.Explanation = fmt.Sprintf("%d-%s placement suggested to grant an %.2f%% affidability.",
									s.Replicas, replicas_s, crit_100)
	}
}
