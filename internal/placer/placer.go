package placer

import (
	"fmt"
	"local/KOrche/core"
	"local/KOrche/internal/api"
	"strings"
)

// -----------------------------------------------------------------------------
// Place()
// Core entrypoint for the placement logic
// -----------------------------------------------------------------------------

func Place(req api.PlacementRequest) (*api.PlacementResponse, error) {
	// 1. Build the cluster
	cluster := core.Cluster{
		Active: make(map[int]*core.WorkerNode),
		Idle:   make(map[int]*core.WorkerNode),
	}

	for _, wnDTO := range req.Cluster.WorkerNodes {
		var node core.WorkerNode = api.BuildWorkerNodeFromDTO(wnDTO)

		// Determine state: active if any resource requested > 0
		if node.IsActive() {
			node.State = core.Active
			cluster.Active[node.AutoID] = &node
		} else {
			node.State = core.Idle
			cluster.Idle[node.AutoID] = &node
		}
	}

	// 2. Build the pod
	var pod core.Pod = api.BuildPodFromDTO(*req.Deployment.Pod)

	// 3. Build scoring function
	hyperparams := core.SetAlgorithmParams(req.Deployment.Algorithm.HyperParams)
	core.ResourceDangerRatio = hyperparams.OvercommitResourceDangerRatio

	scoringCfg := core.ScoringConfig{
		ResourceFit: req.Deployment.Algorithm.ResourceFit,
		Weights:     req.Deployment.Algorithm.Weights,
	}
	scoringFunc, explainScoringFunc := core.BuildScoringFunction(scoringCfg, hyperparams)

	// Correct Formatting of Algorithm.Type
	var requiredAlgo string = strings.ToLower(req.Deployment.Algorithm.Type)
	requiredAlgo = strings.ReplaceAll(requiredAlgo, "_", "")

	cluster.ComputeEnergyConstants(hyperparams) //Sets MaxEnergyCost and CurrEnergyCost

	// 4. Dispatch to the requested algorithm
	var solution core.Solution
	switch requiredAlgo {
	case "greedy":
		solution = core.AddingNewPodGreedy(&cluster, &pod, scoringFunc, explainScoringFunc, hyperparams)
	case "dpstateaware":
		solution = core.AddingNewPodDPStateAware(&cluster, &pod, scoringFunc, explainScoringFunc, hyperparams)
	case "dpstateagnostic":
		solution = core.AddingNewPodDPStateAgnostic(&cluster, &pod, scoringFunc, explainScoringFunc, hyperparams)
	default:
		fmt.Printf("[WARNING] Unrecgnized Algorithm type: %s. Using K8s\n", req.Deployment.Algorithm.Type)
		solution = core.AddingNewPodK8s(&cluster, &pod, scoringFunc, explainScoringFunc, hyperparams)
		// return nil, fmt.Errorf("unknown algorithm type: %s", req.Deployment.Algorithm.Type)
	}

	// 5. Return the solution (to be serialized as response)
	resp := api.BuildResponseFromSolution(solution)
	return &resp, nil
}
