package main

import "fmt"

type Pod struct {
	ID          int
	RealTime    bool
	CPU         BasicResourceType
	Disk        BasicResourceType
	RAM         BasicResourceType
	Criticality Criticality
	replicas    int //to do

	computation_left float32
	// completion_notified		bool
}

//Create
func createRandomPod(id int) *Pod {
	const _RESOURCE_MIN int = 2
	const _RESOURCE_MAX int = 15
	const _RESOURCE_UNIT int = 50
	const _LIM_MAX_RATIO float32 = 2.5

	var rnd = rand_01()
	var rt bool = rnd >= 0.5
	var crit Criticality
	if rt {
		rnd = rand_01()
		if rnd < 2./3. {
			crit = MidCriticality
		} else {
			crit = HighCriticality
		}
	} else {
		rnd = rand_01()
		if rnd < 2./3. {
			crit = LowCriticality
		} else if rnd < 8./9. {
			crit = MidCriticality
		} else {
			crit = HighCriticality
		}
	}
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
	// Computation left
	var cp_left float32 = float32(rand_ab_int(5, int(float32(m)*0.1)))
	// Replicas
	var replicas int = 1
	if crit >= MidCriticality {
		replicas += 2
	}
	if crit >= HighCriticality {
		replicas += 2
	}

	return &Pod{
		ID:               id,
		CPU:              cpu,
		Disk:             disk,
		RAM:              ram,
		RealTime:         rt,
		Criticality:      crit,
		computation_left: cp_left,
		// completion_notified : false,
		replicas: replicas,
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

		replicas: p.replicas,
	}
}

func (p Pod) String() string {
	return fmt.Sprintf("Pod %d.\n\tReal time:\t\t%t\n\tCriticality\t\t%s (%d replicas)\n\tCPU:\t\t\t%s\n\tDisk:\t\t\t%s\n\tRAM:\t\t\t%s\n",
		p.ID, p.RealTime, p.Criticality, p.replicas, p.CPU, p.Disk, p.RAM,
	)
}

/* Run */
const _LIGHT_INTERFERENCE_CHANCE float32 = 0.1
const _HEAVY_INTERFERENCE_CHANCE float32 = 0.01

func (p *Pod) Run(wn *WorkerNode) bool {
	if wn.Assurance == HighAssurance {
		p.computation_left--
	} else {
		r := rand_01()
		d := rand_01()
		if r < _HEAVY_INTERFERENCE_CHANCE {
			p.computation_left -= d
		} else if r < _LIGHT_INTERFERENCE_CHANCE+_HEAVY_INTERFERENCE_CHANCE {
			p.computation_left -= (0.5 * d) + 0.5
		}
	}
	return p.computation_left <= 0
}
