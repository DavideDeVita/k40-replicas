package main

import (
	"log"
)

var _MAX_ENERGY_COST = -1

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

/********************** Other params aware **********************/
/** I always want to minimize this:
  between Idle nodes: obviously
  between Active nodes: I want to add to the least costly, hoping the most costly will shutdown eventually
*/
func _energyCost_ratio(node *WorkerNode, pod *Pod) float32 {
	return float32(node.EnergyCost) / float32(_MAX_ENERGY_COST)
}

func _computationPower_ratio(node *WorkerNode, pod *Pod) float32 {
	// return 1. / float32(node.Computation_Power)
	return 2. / float32(node.Computation_Power+1)
}

func _sigma_assurance(node *WorkerNode, pod *Pod) float32 {
	return float32(1. / (1. + power_f32((node.Assurance)/(1-node.Assurance), 1.5)))
}

func _sigma_assurance_wasteless(node *WorkerNode, pod *Pod) float32 {
	if node.Assurance >= pod.Criticality.value() {
		return 0.
	}
	return _sigma_assurance(node, pod)
}

func _hsigma_assurance(node *WorkerNode, pod *Pod) float32 {
	sigma := _sigma_assurance(node, pod)
	return float32(2.*(1.-node.Assurance)) - sigma
}

func _hsigma_assurance_wasteless(node *WorkerNode, pod *Pod) float32 {
	if node.Assurance >= pod.Criticality.value() {
		return 0.
	}
	return _hsigma_assurance(node, pod)
}

func _log10_assurance(node *WorkerNode, pod *Pod) float32 {
	// return 1. - (-log10_f32(1.-node.Assurance) / 10.)
	a := ((node.Assurance-0.75)/(1-0.75))*(1-0.95) + 0.95
	return 1. - (-log10_f32(1.-a) / 10.)
}

func _log10_assurance_wasteless(node *WorkerNode, pod *Pod) float32 {
	if node.Assurance >= pod.Criticality.value() {
		return 0.
	}
	return _log10_assurance(node, pod)
}

func _expFrac_assurance_wasteless(node *WorkerNode, pod *Pod) float32 {
	if node.Assurance >= pod.Criticality.value() {
		return 0.
	}
	diff := pod.Criticality.value()-node.Assurance
	var steep float32 = 3.33
	return (1-exppower_f32(-steep*diff))/1-exppower_f32(-steep)
}

func _rt_waste(node *WorkerNode, pod *Pod) float32 {
	if node.RealTime && !pod.RealTime {
		return 1.
	}
	return 0.
}

/********************** Multiple things aware **********************/
/**	0	k8s:	basic Kubernetes
1	ec: 	energy cost
2	cp:		computation power
3	la:		look ahead
*/
var _funcs []func(node *WorkerNode, p *Pod) float32
var _weights []float32
var _names []string
var _w_denom float32 = 0.

func init_scoring_params(test Test) {
	_funcs = append([]func(node *WorkerNode, p *Pod) float32{test.Placing_scorer}, test.Multi_obj_funcs...)
	_weights = append([]float32{test.Placing_w}, test.Multi_obj_w...)
	_names = append([]string{"k8s placer"}, test.Multi_obj_names...)
	_w_denom = 0
	for _, w := range _weights {
		_w_denom += w
	}
}

func evaluate_score(node *WorkerNode, p *Pod) float32 {
	var score float32 = 0.
	for i, f := range _funcs {
		if _Log >= Log_Scores {
			log.Printf("Eval score (%s) \t%.2f * %.2f = %.2f\n", _names[i], f(node, p), _weights[i], f(node, p)*_weights[i])
		}
		score += f(node, p) * _weights[i]
	}

	return score / _w_denom
}

/********************** Eval conditions **********************/

// No need to have 3 diff functions here
func k8s_leastAllocated_condition(score float32, best float32) bool {
	return score < best
}

func k8s_mostAllocated_condition(score float32, best float32) bool {
	return score < best
}

func k8s_RequestedToCapacityRatio_condition(score float32, best float32) bool {
	return score < best
}
