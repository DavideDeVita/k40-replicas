package main

import (
	"fmt"
)

type Pod struct {
	ID          int
	RealTime    bool
	CPU         BasicResourceType
	Disk        BasicResourceType
	RAM         BasicResourceType
	Criticality _Criticality

	computation_left float32
	// completion_notified		bool
}

// Create
func createRandomPod(id int) *Pod {
	const _RESOURCE_MIN int = 3
	const _RESOURCE_MAX int = 18
	const _RESOURCE_UNIT int = 50
	const _LIM_MAX_RATIO float32 = 3.

	// const _LOAD_REPLICAS_MIN int = 1
	// const _LOAD_REPLICAS_MAX int = 5

	const _CP_MIN int = 15
	var _CP_MAX_func func(int)int = func(tot_pods int)int {
		if tot_pods>100{
			return int(150.*(log10_int(tot_pods)-2.))+_CP_MIN		// 100.*
		}else{
			return 25
		}
	} 
	// const _CP_MAX_PERC func(int)int = 15 // = m*_CP_MAX_PERC/100

	var rt bool = rand_01() >= 0.5
	var crit _Criticality = random_Criticality(rt)
	// CPU
	var req int = rand_ab_int(_RESOURCE_MIN, _RESOURCE_MAX)
	lim_max := int((float32(req) * _LIM_MAX_RATIO))
	var lim int = rand_ab_int(req, lim_max)
	req *= _RESOURCE_UNIT
	lim *= _RESOURCE_UNIT
	cpu := BasicResourceType{
		request: req,
		limit:   lim,
	}
	//Disk
	req = rand_ab_int(_RESOURCE_MIN, _RESOURCE_MAX)
	lim_max = int((float32(req) * _LIM_MAX_RATIO))
	lim = rand_ab_int(req, lim_max)
	req *= _RESOURCE_UNIT
	lim *= _RESOURCE_UNIT
	disk := BasicResourceType{
		request: req,
		limit:   lim,
	}
	//RAM
	req = rand_ab_int(_RESOURCE_MIN, _RESOURCE_MAX)
	lim_max = int((float32(req) * _LIM_MAX_RATIO))
	lim = rand_ab_int(req, lim_max)
	req *= _RESOURCE_UNIT
	lim *= _RESOURCE_UNIT
	ram := BasicResourceType{
		request: req,
		limit:   lim,
	}
	// Computation left(
	var cp_left float32 = float32(rand_ab_int(_CP_MIN, _CP_MAX_func(m)))
	// if m < 100 {
	// 	cp_left *= 100
	// }

	return &Pod{
		ID:               id,
		CPU:              cpu,
		Disk:             disk,
		RAM:              ram,
		RealTime:         rt,
		Criticality:      crit,
		computation_left: cp_left,
	}
}

func (p Pod) Copy() *Pod {
	return &Pod{
		ID:          p.ID,
		CPU:         p.CPU.Copy(),
		Disk:        p.Disk.Copy(),
		RAM:         p.RAM.Copy(),
		RealTime:    p.RealTime,
		Criticality: p.Criticality,

		computation_left: p.computation_left,
	}
}

func (p Pod) String() string {
	ret := ""
	ret += fmt.Sprintf("Pod %d.\n", p.ID)
	ret += fmt.Sprintf("\tReal time:\t\t%t\n", p.RealTime)
	// ret += fmt.Sprintf("\tCriticality\t\t%s (%d replicas)\n", p.Criticality, p.replicas)
	ret += fmt.Sprintf("\tCriticality\t\t%s\n", p.Criticality)
	ret += fmt.Sprintf("\tCPU:\t\t\t%s\n", p.CPU)
	ret += fmt.Sprintf("\tDisk:\t\t\t%s\n", p.Disk)
	ret += fmt.Sprintf("\tRAM:\t\t\t%s\n", p.RAM)
	ret += fmt.Sprintf("\textimated computation time:\t%.2f\n", p.computation_left)
	return ret
}

/* Run */

func (p *Pod) Run(wn *WorkerNode, interference bool, heavyInterference bool) bool {
	var mult float32 = 1.
	if heavyInterference { // heavy I. means cancel out the entire computation for this iteration
		mult = 0.
	}else if interference { //simple I. means cancel half (?)
		mult = 0.5
	}
	p.computation_left -= float32(wn.Computation_Power) * mult
	// p.computation_left -= 1. * mult
	return p.computation_left <= 0.
}
