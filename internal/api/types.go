package api

import (
	"encoding/json"
)

// -----------------------------------------------------------------------------
// WorkerNode DTOs
// -----------------------------------------------------------------------------

// ResourceStatsDTO mirrors the JSON input for each basic resource (cpu, memory, storage).
type ResourceStatsDTO struct {
	Capacity  string `json:"capacity"`  // total available
	Requested string `json:"requested"` // currently requested
	Limit     string `json:"limit"`     // upper bound (used_limit)
}

// WorkerNodeDTO mirrors the structure expected from JSON input for each node.
type WorkerNodeDTO struct {
	ID             string           `json:"id"`
	CPU            ResourceStatsDTO `json:"cpu"`
	Memory         ResourceStatsDTO `json:"memory"`
	Storage        ResourceStatsDTO `json:"storage"`
	RealTime       bool             `json:"real_time"`
	EnergyCostCoef int              `json:"energy_cost"`
	EnergyCostIdle int              `json:"energy_cost_idle"`
	Assurance      float32          `json:"assurance"`
}

// ClusterDTO wraps the list of nodes in the cluster request.
type ClusterDTO struct {
	WorkerNodes []WorkerNodeDTO `json:"worker_nodes"`
}

// -----------------------------------------------------------------------------
// Pod DTOs
// -----------------------------------------------------------------------------

// PodResourcesDTO mirrors the JSON input for each pod resource.
type PodResourcesDTO struct {
	Requested string `json:"requested"`
	Limit     string `json:"limit"`
}

// PodDTO represents a pod specification as received from the request.
type PodDTO struct {
	ID          string          `json:"id"`
	RealTime    bool            `json:"real_time"`
	CPU         PodResourcesDTO `json:"cpu"`
	Memory      PodResourcesDTO `json:"memory"`
	Storage     PodResourcesDTO `json:"storage"`
	Criticality float32         `json:"criticality"`
}

// -----------------------------------------------------------------------------
// Deployment & Algorithm DTOs
// -----------------------------------------------------------------------------

// AlgorithmConfigDTO defines which algorithm and scoring configuration to use.
type AlgorithmConfigDTO struct {
	Type        string                 `json:"type"`        // e.g. "Greedy", "DP_stateAware", "DP_stateAgnostic"
	ResourceFit string                 `json:"resourceFit"` // e.g. "leastAllocated", "mostAllocated", "requestedToCapacityRatio"
	Weights     map[string]float32     `json:"weights,omitempty"`
	HyperParams map[string]interface{} `json:"hyperparams,omitempty"`
}

// DeploymentDTO groups the job/pod and algorithm specification.
type DeploymentDTO struct {
	Pod       *PodDTO             `json:"pod"`
	Algorithm *AlgorithmConfigDTO `json:"algorithm"`
}

// UnmarshalJSON permette di accettare sia "pod" che "job" come chiavi JSON.
func (d *DeploymentDTO) UnmarshalJSON(data []byte) error {
	// Alias per evitare ricorsione infinita
	type Alias DeploymentDTO
	aux := struct {
		*Alias
		Job *PodDTO `json:"job"`
	}{
		Alias: (*Alias)(d),
	}

	// Primo unmarshal standard
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Se "pod" non è presente ma "job" sì, assegna quest'ultimo
	if d.Pod == nil && aux.Job != nil {
		d.Pod = aux.Job
	}

	// fmt.Printf("[ERROR] Job entry not found in request!\n")

	return nil
}

// -----------------------------------------------------------------------------
// Placement request root DTO
// -----------------------------------------------------------------------------

type PlacementRequest struct {
	Cluster    *ClusterDTO    `json:"cluster"`
	Deployment *DeploymentDTO `json:"deployment"`
}

// -----------------------------------------------------------------------------
// Placement Response
// -----------------------------------------------------------------------------
type PlacementResponse struct {
	PodID		   string             `json:"pod_id"`
	Accepted       bool               `json:"accepted"`
	Replicas       int                `json:"replicas"`
	Probability    float32            `json:"probability,omitempty"`
	Nodes          []PlacementNodeDTO `json:"deploy_on,omitempty"`
	Explanation    string             `json:"explanation"`
	DeltaEnergy    float32            `json:"energy_delta"`
}

type PlacementNodeDTO struct {
	ID        string             `json:"id"`
	CPU       ResourceStatsDTO   `json:"cpu"`
	Memory    ResourceStatsDTO   `json:"memory"`
	Storage   ResourceStatsDTO   `json:"storage"`
	Score     float32            `json:"score"`
	Subscores map[string]float32 `json:"explain"`
}
