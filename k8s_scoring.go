package main

// import "log"

/* weights */
const k8s_w float32 = 2
const ec_w float32 = 1 //energy cost
const cp_w float32 = 1 //energy cost

/********************** K8s Scorer **********************/

func k8s_leastAllocated_score(node *WorkerNode, p *Pod) float32 {
	// Compute the percentage of allocated resources for Basic Resource
	allocCPU_p, allocDisk_p, allocRAM_p := node.AllocatedPercentage()

	// Return the average of Basic Resources alloc percentages
	return (allocCPU_p + allocDisk_p + allocRAM_p) / 3.
}

func k8s_mostAllocated_score(node *WorkerNode, p *Pod) float32 {
	// Compute the percentage of freeated resources for Basic Resource
	freeCPU_p, freeDisk_p, freeRAM_p := node.UnusedPercentage()

	// Return the average of Basic Resources free percentages       NB: I am always searching the min
	return (freeCPU_p + freeDisk_p + freeRAM_p) / 3.
}

func k8s_requestedToCapacityRatio_score(node *WorkerNode, p *Pod) float32 {
	s := node.status
	// Compute the requested-to-capacity ratio for each Basic Resource
	cpuRatio := float32(p.CPU.request) / float32(s.unrequestedCPU)
	DiskRatio := float32(p.Disk.request) / float32(s.unrequestedDisk)
	ramRatio := float32(p.RAM.request) / float32(s.unrequestedRAM)

	var _cpu_weight, _Disk_weight, _ram_weight float32 = 1., 1., 1.

	cpuRatio *= _cpu_weight
	DiskRatio *= _Disk_weight
	ramRatio *= _ram_weight

	// I invert it (1 - ...) because i am always looking for the minimum
	return 1. - ((cpuRatio + DiskRatio + ramRatio) / 3.)
}

/********************** Energy Cost aware **********************/
/** I always want to minimize this:
  between Idle nodes: obviously
  between Active nodes: I want to add to the least costly, hoping the most costly will shutdown eventually
*/
func _energyCost_ratio(node *WorkerNode) float32 {
	return float32(node.EnergyCost) / float32(_MAX_ENERGY_COST)
}

func costAware_leastAllocated_score(node *WorkerNode, p *Pod) float32 {
	var k8s_score float32 = k8s_leastAllocated_score(node, p)
	var ec_score float32 = _energyCost_ratio(node)
	// log.Println("k8s: ", k8s_score, " + ", ec_score)
	//
	return (k8s_score*k8s_w + ec_score*ec_w) / (k8s_w + ec_w)
}

func costAware_mostAllocated_score(node *WorkerNode, p *Pod) float32 {
	var k8s_score float32 = k8s_mostAllocated_score(node, p)
	var ec_score float32 = _energyCost_ratio(node)

	// I always want to minimize the the score
	return (k8s_score*k8s_w + ec_score*ec_w) / (k8s_w + ec_w)
}

func costAware_requestedToCapacityRatio_score(node *WorkerNode, p *Pod) float32 {
	var k8s_score float32 = k8s_requestedToCapacityRatio_score(node, p)
	var ec_score float32 = _energyCost_ratio(node)

	//
	return (k8s_score*k8s_w + ec_score*ec_w) / (k8s_w + ec_w)
}

/********************** Look Ahead K8s Scorer **********************/

func la_leastAllocated_score(node *WorkerNode, p *Pod) float32 {
	// Compute the percentage of allocated resources for Basic Resource
	allocCPU_p, allocDisk_p, allocRAM_p := node.AllocatedPercentage()
	//Look ahead
	allocCPU_p += float32(p.CPU.request) / float32(node.CPU_Capacity)
	allocDisk_p += float32(p.Disk.request) / float32(node.Disk_Capacity)
	allocRAM_p += float32(p.RAM.request) / float32(node.RAM_Capacity)

	// Return the average of Basic Resources alloc percentages
	return (allocCPU_p + allocDisk_p + allocRAM_p) / 3.
}

func la_mostAllocated_score(node *WorkerNode, p *Pod) float32 {
	// Compute the percentage of freeated resources for Basic Resource
	freeCPU_p, freeDisk_p, freeRAM_p := node.UnusedPercentage()
	//Look ahead
	freeCPU_p -= float32(p.CPU.request) / float32(node.CPU_Capacity)
	freeDisk_p -= float32(p.Disk.request) / float32(node.Disk_Capacity)
	freeRAM_p -= float32(p.RAM.request) / float32(node.RAM_Capacity)

	// Return the average of Basic Resources free percentages       NB: I am always searching the min
	return (freeCPU_p + freeDisk_p + freeRAM_p) / 3.
}

/* There is no LA_requestedToCapcityRatio_score */

/********************** Multiple things aware **********************/
/**	0	k8s:	basic Kubernetes
1	ec: 	energy cost
2	cp:		computation power
3	la:		look ahead
*/
var _k8s_func func(node *WorkerNode, p *Pod) float32
var _lookAhead_k8s_func func(node *WorkerNode, p *Pod) float32
var _active_vec []bool
var _multi_w []float32
var _w_denom float32 = 0.

func init_multiAware_params(params MultiAwareParams) {
	_k8s_func = params.k8s_func
	_lookAhead_k8s_func = params.la_k8s_func
	_active_vec = params.active
	_multi_w = params.weights
	for i := 0; i < 4; i++ {
		if _active_vec[i] {
			_w_denom += _multi_w[i]
		}
	}
}

func _computationPower_ratio(node *WorkerNode, pod *Pod) float32 {
	return 1. / float32(node.Computation_Power)
}

func parametric_scores_aware_score(node *WorkerNode, p *Pod) float32 {
	var k8s_score float32 = _k8s_func(node, p)
	var ec_score float32 = _energyCost_ratio(node)
	var cp_score float32 = _computationPower_ratio(node, p)
	var la_score float32 = 0
	if _active_vec[3] { //Check cause it could be nil
		la_score = _lookAhead_k8s_func(node, p)
	}
	// log.Println("k8s: ", k8s_score, " + ", ec_score)
	//
	return (k8s_score*_multi_w[0] + ec_score*_multi_w[1] + cp_score*_multi_w[2] + la_score*_multi_w[3]) / _w_denom
}

/********************** Eval conditions **********************/

//No need to have 3 diff functions here
func k8s_leastAllocated_condition(score float32, best float32) bool {
	return score < best
}

func k8s_mostAllocated_condition(score float32, best float32) bool {
	return score < best
}

func k8s_RequestedToCapacityRatio_condition(score float32, best float32) bool {
	return score < best
}
