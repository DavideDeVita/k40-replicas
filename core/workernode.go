package core

var _NEXT_AUTOID = 0

// ResourceStats describes the usage of a basic resource type (cpu, memory, storage).
type ResourceStats struct {
	Capacity    int `json:"capacity"`  // total available
	Requested   int `json:"requested"` // currently requested
	Limit       int `json:"limit"`     // upper bound (used_limit)
	Unrequested int `json:"-"`         // derived: Capacity - Requested
	UnusedLimit int `json:"-"`         // derived: Capacity - Limit
}

// ComputeDerived calculates the derived fields for Unrequested and UnusedLimit.
func (r *ResourceStats) ComputeDerived() {
	r.Unrequested = r.Capacity - r.Requested
	r.UnusedLimit = r.Capacity - r.Limit
}

func (r ResourceStats) isInUse() bool {
	return r.Requested > 0 || r.Limit > 0
}

// WorkerNode represents a cluster node with its resources and metadata.
type WorkerNode struct {
	AutoID         int           `json:"-"`
	ID             string        `json:"id"`
	CPU            ResourceStats `json:"cpu"`
	Memory         ResourceStats `json:"memory"`
	Storage        ResourceStats `json:"storage"`
	RealTime       bool          `json:"real_time"`
	EnergyCostCoef int           `json:"energy_cost_coef"`
	EnergyCostIdle int           `json:"energy_cost_idle"`
	Assurance      float32       `json:"assurance"`

	State ClusterNodeState `json:"-"` // Active / Idle
}

// NextAutoID restituisce un nuovo ID interno incrementale.
func NextAutoID() int {
	_NEXT_AUTOID++
	return _NEXT_AUTOID - 1
}

// Getters
func (node *WorkerNode) UnrequestedPercentage() (float32, float32, float32) {
	// Compute the percentage of unallocated resources for Basic Resource
	freeCPU_p := float32(node.CPU.Unrequested) / float32(node.CPU.Capacity)
	freeMemory_p := float32(node.Memory.Unrequested) / float32(node.Memory.Capacity)
	freeStorage_p := float32(node.Storage.Unrequested) / float32(node.Storage.Capacity)

	// Return the average of Basic Resources free percentages
	return freeCPU_p, freeMemory_p, freeStorage_p
}
func (node *WorkerNode) RequestedPercentage() (float32, float32, float32) {
	// Compute the percentage of allocated resources for Basic Resource
	allocCPU_p := float32(node.CPU.Requested) / float32(node.CPU.Capacity)
	allocMemory_p := float32(node.Memory.Requested) / float32(node.Memory.Capacity)
	allocStorage_p := float32(node.Storage.Requested) / float32(node.Storage.Capacity)

	// Return the average of Basic Resources free percentages
	return allocCPU_p, allocMemory_p, allocStorage_p
}

func (node *WorkerNode) UnusedLimitPercentage() (float32, float32, float32) {
	// Compute the percentage of unallocated resources for Basic Resource
	freeCPU_p := float32(node.CPU.UnusedLimit) / float32(node.CPU.Capacity)
	freeMemory_p := float32(node.Memory.UnusedLimit) / float32(node.Memory.Capacity)
	freeStorage_p := float32(node.Storage.UnusedLimit) / float32(node.Storage.Capacity)

	// Return the average of Basic Resources free percentages
	return freeCPU_p, freeMemory_p, freeStorage_p
}
func (node *WorkerNode) UsedLimitPercentage() (float32, float32, float32) {
	// Compute the percentage of allocated resources for Basic Resource
	allocCPU_p := float32(node.CPU.Limit) / float32(node.CPU.Capacity)
	allocMemory_p := float32(node.Memory.Limit) / float32(node.Memory.Capacity)
	allocStorage_p := float32(node.Storage.Limit) / float32(node.Storage.Capacity)

	// Return the average of Basic Resources free percentages
	return allocCPU_p, allocMemory_p, allocStorage_p
}

// Eligibility
func (wn WorkerNode) baseRequirementsMatch(p *Pod) bool {
	return wn.CPU.Unrequested >= p.CPU.Requested &&
		wn.Memory.Unrequested >= p.Memory.Requested &&
		wn.Storage.Unrequested >= p.Storage.Requested
}
func (wn WorkerNode) advancedRequirementsMatch(p *Pod) bool {
	return true
}
func (wn *WorkerNode) EligibleFor(pod *Pod) bool {
	return (wn.RealTime || !pod.RealTime) &&
		wn.advancedRequirementsMatch(pod) &&
		wn.baseRequirementsMatch(pod)
}

// Check State
func (wn WorkerNode) IsActive() bool {
	return wn.CPU.isInUse() || wn.Memory.isInUse() || wn.Storage.isInUse()
}

func (wn WorkerNode) ComputeLinearEnergyCost(withPod *Pod) float32 {
	// Avg tra Requested e Limit della CPU dopo l'inserimento di questo Pod
	usage := float32(wn.CPU.Requested + wn.CPU.Limit)
	cap := float32(wn.CPU.Capacity)

	// Hypotetical new pod added
	if withPod != nil {
		usage += float32(withPod.CPU.Requested + withPod.CPU.Limit)
	}
	usage /= 2.

	p_load := usage / cap
	if p_load > 1 {
		p_load = 1
	}
	// Compute ret
	ret := p_load * float32(wn.EnergyCostCoef)
	if ret < float32(wn.EnergyCostIdle) {
		return float32(wn.EnergyCostIdle)
	}
	return ret
}

func (wn WorkerNode) ComputeAbsoluteEnergyCost() float32 {
	if wn.IsActive() {
		return float32(wn.EnergyCostCoef)
	} else {
		return float32(wn.EnergyCostIdle)
	}
}
