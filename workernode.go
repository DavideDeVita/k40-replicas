package main

import (
	"fmt"
	"os"
)

const _REALISTIC_OVERPRICE bool = true

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
	ID            int
	CPU_Capacity  int
	Disk_Capacity int
	RAM_Capacity  int
	RealTime      bool
	EnergyCost    int
	Assurance     Assurance

	pods   map[int]*Pod
	status WN_Status
}

// Create
func createRandomWorkerNode(id int) WorkerNode {
	const br_unit, br_min, br_max int = 100, 15, 50
	const cost_unit, cost_min, cost_max int = 50, 5, 50

	var rt = rand_01() >= 0.5
	var cpu_capacity int = rand_ab_int(br_min, br_max) * br_unit
	var disk_capacity int = rand_ab_int(br_min, br_max) * br_unit
	var ram_capacity int = rand_ab_int(br_min, br_max) * br_unit
	var cost_f float32 = float32(rand_ab_int(cost_min, cost_max) * cost_unit)

	var assurance Assurance
	var r = rand_01()
	if r >= 0.5 || (rt && r >= 1./3.) {
		assurance = HighAssurance
	} else {
		assurance = LowAssurance
	}

	/*Realistic overprice*/
	if _REALISTIC_OVERPRICE {
		if rt {
			cost_f *= 1.+rand_01()
		}
		if assurance == HighAssurance {
			cost_f *= 1.+rand_01()*1.75
		}
		avg_br := (br_min+br_max)/2. * br_unit
		if cpu_capacity > avg_br{
			cost_f *= 1+rand_01()*0.5
		}
		if disk_capacity > avg_br{
			cost_f *= 1+rand_01()*0.5
		}
		if ram_capacity > avg_br{
			cost_f *= 1+rand_01()*0.5
		}
	}

	var cost_i int = int(cost_f/float32(cost_unit)) * cost_unit
	/*UPDATE GLOBAL MAX_ENERGY_COST*/
	if _MAX_ENERGY_COST < cost_i {
		_MAX_ENERGY_COST = cost_i
	}

	return WorkerNode{
		ID:            id,
		CPU_Capacity:  cpu_capacity,
		Disk_Capacity: disk_capacity,
		RAM_Capacity:  ram_capacity,
		RealTime:      rt,
		EnergyCost:    cost_i,

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

		Assurance: wn.Assurance,

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
	return fmt.Sprintf("Worker Node %d (%d Pods currenty deployed).\n\tReal time:\t\t%t\n\tCPU capacity:\t\t%d\tr:%d\tu:%d\n\tDisk capacity:\t\t%d\tr:%d\tu:%d\n\tRAM capacity:\t\t%d\tr:%d\tu:%d\n\tActivation cost\t\t%d\n\tAssurance \t%s\n",
		wn.ID, len(wn.pods), wn.RealTime, wn.CPU_Capacity, wn.status.requestedCPU, wn.status.unrequestedCPU, wn.Disk_Capacity, wn.status.requestedDisk, wn.status.unrequestedDisk, wn.RAM_Capacity, wn.status.requestedRAM, wn.status.unrequestedRAM, wn.EnergyCost, wn.Assurance,
	)
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
		fmt.Printf("Pod %d already in worker node %d\n", pod.ID, wn.ID)
		os.Exit(1)
	}
	wn.pods[pod.ID] = pod
	wn.status.PodWasAdded(pod)
	// fmt.Printf("Pod %d inserted in Worker Node %d\n", pod.ID, wn.ID)
	// fmt.Println(wn)
}

func (wn *WorkerNode) RemovePod(pod *Pod) {
	if _, exists := wn.pods[pod.ID]; exists {
		delete(wn.pods, pod.ID)
		wn.status.PodWasRemoved(pod)
		return
	}
	fmt.Printf("Pod %d not in worker node %d\n", pod.ID, wn.ID)
	os.Exit(1)
	// fmt.Printf("Pod %d inserted in Worker Node %d\n", pod.ID, wn.ID)
	// fmt.Println(wn)
}

func (wn *WorkerNode) RunPods() bool {
	if len(wn.pods) > 0 {
		completed := make([]*Pod, 0)
		for _, pod := range wn.pods {
			complete := pod.Run(wn)
			if complete {
				if _Log >= Log_Some {
					fmt.Printf("Pod %d completed on worker node %d\n", pod.ID, wn.ID)
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
