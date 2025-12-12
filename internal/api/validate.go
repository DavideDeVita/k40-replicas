package api

import (
	"fmt"
)

// -----------------------------------------------------------------------------
// PlacementRequest Validation
// -----------------------------------------------------------------------------
func (r *PlacementRequest) Validate() error {
	if r.Cluster == nil {
		return fmt.Errorf("missing 'cluster' section in request")
	}
	if len(r.Cluster.WorkerNodes) == 0 {
		return fmt.Errorf("cluster must contain at least one 'worker_node'")
	}
	if err := r.Cluster.Validate(); err != nil {
		return err
	}

	if r.Deployment == nil {
		return fmt.Errorf("missing 'deployment' section in request")
	}
	if err := r.Deployment.Validate(); err != nil {
		return err
	}

	if r.Deployment.Algorithm.OutputsAmount < 1 {
		fmt.Printf("[WARN] 'OutputsAmount' must be at least 1 but found %d, converted to 1\n", r.Deployment.Algorithm.OutputsAmount)
		r.Deployment.Algorithm.OutputsAmount = 1
	}

	return nil
}

// -----------------------------------------------------------------------------
// Cluster Validation
// -----------------------------------------------------------------------------

func (c *ClusterDTO) Validate() error {
	if len(c.WorkerNodes) == 0 {
		return fmt.Errorf("cluster must contain at least one worker node")
	}
	for i, wn := range c.WorkerNodes {
		if err := wn.Validate(); err != nil {
			return fmt.Errorf("worker_nodes[%d]: %v", i, err)
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Worker Node Validation
// -----------------------------------------------------------------------------

func (w *WorkerNodeDTO) Validate() error {
	if w.ID == "" {
		return fmt.Errorf("missing 'id' for worker node")
	}
	if err := validateResourceStats("cpu", w.CPU); err != nil {
		return err
	}
	if err := validateResourceStats("memory", w.Memory); err != nil {
		return err
	}
	if err := validateResourceStats("storage", w.Storage); err != nil {
		return err
	}
	if w.EnergyCostCoef <= 0 {
		return fmt.Errorf("invalid or missing 'energy_cost' for node %q.\t'energy_cost' must be > 0", w.ID)
	}
	if w.EnergyCostIdle < 0 {
		return fmt.Errorf("invalid 'energy_cost_idle' for node %q.\t'energy_cost_idle' must be >= 0", w.ID)
	}
	if w.Assurance <= 0 || w.Assurance > 1 {
		return fmt.Errorf("invalid or missing 'assurance' for node %q\t'assurance' must be  must be in (0,1]", w.ID)
	}
	return nil
}

func validateResourceStats(prefix string, r ResourceStatsDTO) error {
	if r.Capacity == "" {
		return fmt.Errorf("missing '%s.capacity'", prefix)
	}
	if r.Requested == "" {
		return fmt.Errorf("missing '%s.requested'", prefix)
	}
	if r.Limit == "" {
		return fmt.Errorf("missing '%s.limit'", prefix)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Deployment Validation
// -----------------------------------------------------------------------------

func (d *DeploymentDTO) Validate() error {
	if d.Pod == nil {
		return fmt.Errorf("missing 'pod' section inside 'deployment'")
	}
	if err := d.Pod.Validate(); err != nil {
		return fmt.Errorf("pod: %v", err)
	}

	if d.Algorithm == nil {
		return fmt.Errorf("missing 'algorithm' section inside 'deployment'")
	}
	if err := d.Algorithm.Validate(); err != nil {
		return fmt.Errorf("algorithm: %v", err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Pod Validation
// -----------------------------------------------------------------------------

func (p *PodDTO) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("missing 'pod.id'")
	}
	if err := validatePodResources("cpu", p.CPU); err != nil {
		return err
	}
	if err := validatePodResources("memory", p.Memory); err != nil {
		return err
	}
	if err := validatePodResources("storage", p.Storage); err != nil {
		return err
	}
	if p.Criticality <= 0 || p.Criticality > 1 {
		return fmt.Errorf("'criticality' must be in (0,1]")
	}
	return nil
}

func validatePodResources(prefix string, r PodResourcesDTO) error {
	if r.Requested == "" {
		return fmt.Errorf("missing '%s.requested'", prefix)
	}
	if r.Limit == "" {
		return fmt.Errorf("missing '%s.limit'", prefix)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Algorithm Config Validation
// -----------------------------------------------------------------------------

func (a *AlgorithmConfigDTO) Validate() error {
	if a.Type == "" {
		return fmt.Errorf("missing 'algorithm.type'")
	}
	// Optional: enforce known types
	validTypes := map[string]bool{
		"Greedy":           true,
		"DP_StateAware":    true,
		"DP_StateAgnostic": true,
		"K8s":              true,
	}
	if !validTypes[a.Type] {
		return fmt.Errorf("unknown algorithm type: %q", a.Type)
	}
	return nil
}
