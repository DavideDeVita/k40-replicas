package main

// import "fmt"

/* weights */
const k8s_w float32 = 0.75
const ec_w float32 = 0.25 //energy cost

/********************** Scorer **********************/

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
	diskRatio := float32(p.Disk.request) / float32(s.unrequestedDisk)
	ramRatio := float32(p.RAM.request) / float32(s.unrequestedRAM)

	var _cpu_weight, _disk_weight, _ram_weight float32 = 1., 1., 1.

	cpuRatio *= _cpu_weight
	diskRatio *= _disk_weight
	ramRatio *= _ram_weight

	// I invert it (1 - ...) because i am always looking for the minimum
	return 1. - ((cpuRatio + diskRatio + ramRatio) / 3.)
}

/********************** Energy Cost aware **********************/
/** I always want to minimize this:
  between Idle nodes: obviously
  between Active nodes: I want to add to the least costly, hoping the most costly will shutdown eventually
*/
func energyCost_ratio(node *WorkerNode) float32 {
	return float32(node.EnergyCost) / float32(_MAX_ENERGY_COST)
}

func costAware_leastAllocated_score(node *WorkerNode, p *Pod) float32 {
	var k8s_score float32 = k8s_leastAllocated_score(node, p)
	var ec_score float32 = energyCost_ratio(node)
	// fmt.Println("k8s: ", k8s_score, " + ", ec_score)
	//
	return k8s_score*k8s_w + ec_score*ec_w
}

func costAware_mostAllocated_score(node *WorkerNode, p *Pod) float32 {
	var k8s_score float32 = k8s_mostAllocated_score(node, p)
	var ec_score float32 = energyCost_ratio(node)

	// I always want to minimize the the score
	return k8s_score*k8s_w + ec_score*ec_w
}

func costAware_requestedToCapacityRatio_score(node *WorkerNode, p *Pod) float32 {
	var k8s_score float32 = k8s_requestedToCapacityRatio_score(node, p)
	var ec_score float32 = energyCost_ratio(node)

	//
	return k8s_score*k8s_w + ec_score*ec_w
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
