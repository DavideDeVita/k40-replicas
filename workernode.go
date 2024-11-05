package main

import (
	"fmt"
	"log"
	"os"
)

const _REALISTIC_OVERPRICE bool = true
var _NumRT int = 0
var _noRT_since float32 = 0. // I use this to 'ensure' enough rt nodes

/*********************** Worker Node Status ***********************/
type WN_Status struct {
	requestedCPU  int
	requestedDisk int
	requestedRAM  int

	unrequestedCPU  int
	unrequestedDisk int
	unrequestedRAM  int
}

func (s *WN_Status) PodWasAdded(p *Pod) {
	s.requestedCPU += p.CPU.request
	s.requestedDisk += p.Disk.request
	s.requestedRAM += p.RAM.request

	s.unrequestedCPU -= p.CPU.request
	s.unrequestedDisk -= p.Disk.request
	s.unrequestedRAM -= p.RAM.request
}

func (s *WN_Status) PodWasRemoved(p *Pod) {
	s.requestedCPU -= p.CPU.request
	s.requestedDisk -= p.Disk.request
	s.requestedRAM -= p.RAM.request

	s.unrequestedCPU += p.CPU.request
	s.unrequestedDisk += p.Disk.request
	s.unrequestedRAM += p.RAM.request
}

/*********************** Worker Node ***********************/
type WorkerNode struct {
	ID                int
	CPU_Capacity      int
	Disk_Capacity     int
	RAM_Capacity      int
	RealTime          bool
	EnergyCost        int
	Assurance         float64    // Assurance is now "the probability of not failing and is to be in [10^-7; 10^-2]"
	Computation_Power int

	pods   map[int]*Pod
	status WN_Status
}

// Create
func createRandomWorkerNode(id int) WorkerNode {
	const br_unit, br_min, br_max int = 100, 15, 50
	const cost_unit, cost_min, cost_max int = 50, 5, 50
	const cp_min, cp_max int = 1, 5

	var p_rt float32 = (_noRT_since+1)/(_noRT_since+2) 
	var rt = rand_01() <= p_rt
	var cpu_capacity int = rand_ab_int(br_min, br_max) * br_unit
	var disk_capacity int = rand_ab_int(br_min, br_max) * br_unit
	var ram_capacity int = rand_ab_int(br_min, br_max) * br_unit
	var cp int = rand_ab_int(cp_min, cp_max)
	var cost_f float32 = float32(rand_ab_int(cost_min, cost_max) * cost_unit)

	var assurance float64
	if rt {
		assurance = rand_10pow(-7, -2)
	} else { //Useless else now
		assurance = rand_10pow(-7, -2)
	}

	/*Realistic overprice*/
	if _REALISTIC_OVERPRICE {
		if rt {
			cost_f *= rand_ab_float(1.25, 2.5)
		}
		if assurance >= 0.9995 {
			cost_f *= rand_ab_float(1., 2.)
		}

		avg_br := (br_min + br_max) / 2. * br_unit
		if cpu_capacity > avg_br {
			cost_f *= rand_ab_float(1., 1.1)
		}
		if disk_capacity > avg_br {
			cost_f *= rand_ab_float(1., 1.1)
		}
		if ram_capacity > avg_br {
			cost_f *= rand_ab_float(1., 1.1)
		}

		if cp > (cp_max+cp_min)/2 {
			cost_f *= rand_ab_float(1., 1.75)
		}
	}

	var cost_i int = int(cost_f/float32(cost_unit)) * cost_unit
	/*UPDATE GLOBAL MAX_ENERGY_COST*/
	if _MAX_ENERGY_COST < cost_i {
		_MAX_ENERGY_COST = cost_i
	}

	if rt{
		_NumRT++
		_noRT_since--
		if _noRT_since<0.{ _noRT_since=0. }
	}else{
		_noRT_since++
	}

	return WorkerNode{
		ID:            id,
		CPU_Capacity:  cpu_capacity,
		Disk_Capacity: disk_capacity,
		RAM_Capacity:  ram_capacity,
		RealTime:      rt,
		EnergyCost:    cost_i,

		Computation_Power: cp,

		Assurance: assurance,

		pods: make(map[int]*Pod),
		status: WN_Status{unrequestedCPU: cpu_capacity,
			unrequestedDisk: disk_capacity,
			unrequestedRAM:  ram_capacity,
		},
	}
}

// Base
func (wn WorkerNode) Copy() *WorkerNode {
	return &WorkerNode{
		ID:            wn.ID,
		CPU_Capacity:  wn.CPU_Capacity,
		Disk_Capacity: wn.Disk_Capacity,
		RAM_Capacity:  wn.RAM_Capacity,
		RealTime:      wn.RealTime,
		EnergyCost:    wn.EnergyCost,

		Assurance:         wn.Assurance,
		Computation_Power: wn.Computation_Power,

		pods: make(map[int]*Pod),
		status: WN_Status{requestedCPU: wn.status.requestedCPU,
			requestedDisk: wn.status.requestedDisk,
			requestedRAM:  wn.status.requestedRAM,

			unrequestedCPU:  wn.status.unrequestedCPU,
			unrequestedDisk: wn.status.unrequestedDisk,
			unrequestedRAM:  wn.status.unrequestedRAM,
		},
	}
}

func (wn WorkerNode) String() string {
	ret := ""
	ret += fmt.Sprintf("Worker Node %d (%d Pods currenty deployed).\n", wn.ID, len(wn.pods))
	ret += fmt.Sprintf("\tReal time:\t\t%t\n", wn.RealTime)
	ret += fmt.Sprintf("\tCPU capacity:\t\t%d\tr:%d\tu:%d\n", wn.CPU_Capacity, wn.status.requestedCPU, wn.status.unrequestedCPU)
	ret += fmt.Sprintf("\tDisk capacity:\t\t%d\tr:%d\tu:%d\n", wn.Disk_Capacity, wn.status.requestedDisk, wn.status.unrequestedDisk)
	ret += fmt.Sprintf("\tRAM capacity:\t\t%d\tr:%d\tu:%d\n", wn.RAM_Capacity, wn.status.requestedRAM, wn.status.unrequestedRAM)
	ret += fmt.Sprintf("\tActivation cost\t\t%d\n", wn.EnergyCost)
	ret += fmt.Sprintf("\tAssurance\t\t%s\n", wn.Assurance)
	ret += fmt.Sprintf("\tComputation power \t%.d\n", wn.Computation_Power)

	return ret
}

// Getters
func (node *WorkerNode) UnusedPercentage() (float32, float32, float32) {
	s := node.status
	// Compute the percentage of unallocated resources for Basic Resource
	freeCPU_p := float32(s.unrequestedCPU) / float32(node.CPU_Capacity)
	freeDisk_p := float32(s.unrequestedDisk) / float32(node.Disk_Capacity)
	freeRAM_p := float32(s.unrequestedRAM) / float32(node.RAM_Capacity)

	// Return the average of Basic Resources free percentages
	return freeCPU_p, freeDisk_p, freeRAM_p
}
func (node *WorkerNode) AllocatedPercentage() (float32, float32, float32) {
	s := node.status
	// Compute the percentage of allocated resources for Basic Resource
	allocCPU_p := float32(s.requestedCPU) / float32(node.CPU_Capacity)
	allocDisk_p := float32(s.requestedDisk) / float32(node.Disk_Capacity)
	allocRAM_p := float32(s.requestedRAM) / float32(node.RAM_Capacity)

	// Return the average of Basic Resources alloc percentages
	return allocCPU_p, allocDisk_p, allocRAM_p
}

// Eligibility
func (wn WorkerNode) baseRequirementsMatch(p *Pod) bool {
	return (wn.RealTime || !p.RealTime) &&
		wn.status.unrequestedCPU >= p.CPU.request &&
		wn.status.unrequestedDisk >= p.Disk.request &&
		wn.status.unrequestedRAM >= p.RAM.request
}
func (wn WorkerNode) advancedRequirementsMatch(p *Pod) bool {
	return true
}
func (wn *WorkerNode) EligibleFor(pod *Pod) bool {
	return (wn.RealTime || !pod.RealTime) &&
		wn.advancedRequirementsMatch(pod) &&
		wn.baseRequirementsMatch(pod)
}

// Actions
func (wn *WorkerNode) InsertPod(pod *Pod) {
	if _, exists := wn.pods[pod.ID]; exists {
		log.Printf("Pod %d already in worker node %d\n", pod.ID, wn.ID)
		os.Exit(1)
	}
	wn.pods[pod.ID] = pod
	wn.status.PodWasAdded(pod)
	// log.Printf("Pod %d inserted in Worker Node %d\n", pod.ID, wn.ID)
	// log.Println(wn)
}

func (wn *WorkerNode) RemovePod(pod *Pod) {
	if _, exists := wn.pods[pod.ID]; exists {
		delete(wn.pods, pod.ID)
		wn.status.PodWasRemoved(pod)
		return
	}
	log.Printf("Pod %d not in worker node %d\n", pod.ID, wn.ID)
	os.Exit(1)
	// log.Printf("Pod %d inserted in Worker Node %d\n", pod.ID, wn.ID)
	// log.Println(wn)
}

// Pud running
func (wn *WorkerNode) RunPods() bool {
	r := float64(rand_01())
	var interference bool = r > wn.Assurance // there is interference if randomValue is greater than assurance (assurance is chance of not having interference)
	
	if len(wn.pods) > 0 {
		completed := make([]*Pod, 0)
		for _, pod := range wn.pods {
			complete := pod.Run(wn, interference)
			if complete {
				if _Log >= Log_Some {
					log.Printf("Pod %d completed on worker node %d\n", pod.ID, wn.ID)
				}
				completed = append(completed, pod)
			}
		}

		for _, pod := range completed {
			wn.RemovePod(pod)
		}
		return len(wn.pods) == 0
	}
	return false
}
