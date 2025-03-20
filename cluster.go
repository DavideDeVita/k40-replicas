package main

import (
	"fmt"
	"log"
	"os"
)

type ClusterNodeState int

const (
	Idle   ClusterNodeState = 0
	Active ClusterNodeState = 1
)

func (s ClusterNodeState) String() string {
	if s == Idle {
		return "Idle"
	} else {
		return "Active"
	}
}

/******************************* Cluster *******************************/
// Cluster struct with Active and Idle, both of type ByAssurance
type Cluster struct {
	name   string
	Active map[int]*WorkerNode
	Idle   map[int]*WorkerNode

	accepted              int
	energeticCost         int
	_Total_Pods           int
	_Total_Energetic_Cost int
}

// NewCluster initializes a Cluster with empty Active and Idle maps for both High and Low
func NewCluster(name string) *Cluster {
	return &Cluster{
		name:   name,
		Active: make(map[int]*WorkerNode),
		Idle:   make(map[int]*WorkerNode),

		accepted:              0,
		energeticCost:         0,
		_Total_Pods:           0,
		_Total_Energetic_Cost: 0,
	}
}

// Helper method, get map based on Assurance
func (c *Cluster) byState(state ClusterNodeState) map[int]*WorkerNode {
	if state == Idle {
		return c.Idle
	} else if state == Active {
		return c.Active
	}
	log.Printf("Errore in cluster.go byState: (%d)\n", state)
	os.Exit(1)
	return nil
}

func _state_String(byState map[int]*WorkerNode) string {
	result := ""
	var rts string = ""
	for id, node := range byState {
		rts = ""
		if node.RealTime {
			rts = "*"
		}
		result += fmt.Sprintf(" %d%s(%d),", id, rts, len(node.pods))
	}
	return result
}

func (c Cluster) String() string {
	result := fmt.Sprintf("%s\n", c.name)
	result += fmt.Sprintf("Pod accepted: %d/%d (%.2f%%)\n", c.accepted, c._Total_Pods, 100.*float32(c.accepted)/float32(c._Total_Pods))
	result += fmt.Sprintf("Energetic cost: %d/%d (%.2f%%)\n", c.energeticCost, c._Total_Energetic_Cost, 100.*float32(c.energeticCost)/float32(c._Total_Energetic_Cost))
	//else
	result += fmt.Sprintf("\tActive Workers: %d\n", len(c.Active))
	result += "\t\t" + _state_String(c.Active) + "\n"
	result += fmt.Sprintf("\tIdle Workers: %d\n", len(c.Idle))
	result += "\t\t" + _state_String(c.Idle) + "\n"
	return result
}

// addWorkerNode adds a WorkerNode to the Idle map, based on the assurance level
func (c *Cluster) AddWorkerNode(wn *WorkerNode) {
	idleMap := c.Idle
	activeMap := c.Active

	_, exists_I := idleMap[wn.ID]
	_, exists_A := activeMap[wn.ID]
	if !exists_I && !exists_A {
		idleMap[wn.ID] = wn
		c._Total_Energetic_Cost += wn.EnergyCost
		// log.Printf("Added WorkerNode %d to Idle\n\n", wn.ID)
	} else {
		log.Printf("WorkerNode %d already exists in Active (%t) or Idle (%t)\n", wn.ID, exists_A, exists_I)
		os.Exit(1)
	}
}

func (c *Cluster) All_list() []*WorkerNode {
	var ret []*WorkerNode = make([]*WorkerNode, 0)
	for _, wn := range c.Active {
		ret = append(ret, wn)
	}
	for _, wn := range c.Idle {
		ret = append(ret, wn)
	}
	return ret
}

func (c *Cluster) All_map() map[int]*WorkerNode {
	var ret map[int]*WorkerNode = make(map[int]*WorkerNode)
	for _, wn := range c.All_list() {
		ret[wn.ID] = wn
	}
	return ret
}

func (c *Cluster) Get_by_Id(id int) (*WorkerNode, ClusterNodeState) {
	var node *WorkerNode
	var exists bool

	node, exists = c.Active[id]
	if exists {
		return node, Active
	}

	node, exists = c.Idle[id]
	if exists {
		return node, Idle
	}
	log.Println("Errore in get By Id", id)
	return nil, 0
}

// activateWorkerNode moves a WorkerNode from Idle to Active
func (c *Cluster) ActivateWorkerNode(id int) {
	idleMap := c.Idle
	activeMap := c.Active

	if wn, idleExists := idleMap[id]; idleExists {
		if _, activeExists := activeMap[id]; !activeExists {
			delete(idleMap, id)
			activeMap[id] = wn
			c.energeticCost += wn.EnergyCost
			// log.Printf("Updated en cost for cluster %s: %d\n", c.name, c.energeticCost)
			if _Log >= Log_All {
				// log.Printf("WorkerNode %d moved from Idle to Active\n", id)
			}
		} else {
			log.Printf("WorkerNode %d is already Active\n", id)
		}
	} else {
		log.Printf("WorkerNode %d not found in Idle\n", id)
	}
}

// deactivateWorkerNode moves a WorkerNode from Active to Idle
func (c *Cluster) DeactivateWorkerNode(id int) {
	idleMap := c.Idle
	activeMap := c.Active

	if wn, idleExists := activeMap[id]; idleExists {
		if _, idleExists := idleMap[id]; !idleExists {
			delete(activeMap, id)
			idleMap[id] = wn
			c.energeticCost -= wn.EnergyCost
			if _Log >= Log_Some {
				log.Printf("WorkerNode %d moved from Active to Idle\n", id)
			}
			// log.Printf("Updated en cost for cluster %s: %d\n", c.name, c.energeticCost)
		} else {
			log.Printf("WorkerNode %d is already Idle\n", id)
		}
	} else {
		log.Printf("WorkerNode %d not found in Active\n", id)
	}
}

func (c *Cluster) PodAccepted() {
	c._Total_Pods++
	c.accepted++
}

func (c *Cluster) PodRejected() {
	c._Total_Pods++
}

/******************************* Solution *******************************/
/* Solutions are just like clusters, I am using a different name just to avoid confusion.
Solutions are plans to insert a new pod, with its replicas.
Nodes in Active are nodes to insert it into,
while nodes in Idle are nodes that need to be woken up, and then inserted into
*/
type Solution struct {
	pod        *Pod
	rejected   bool
	Active     map[int]*WorkerNode
	Idle       map[int]*WorkerNode
	n_replicas int
}

// NewSolution initializes a Solution with empty Active and Idle maps for both High and Low
func NewSolution(pod *Pod) Solution {
	return Solution{
		pod:        pod,
		rejected:   false,
		Active:     make(map[int]*WorkerNode),
		Idle:       make(map[int]*WorkerNode),
		n_replicas: 0,
	}
}

func (s *Solution) Reject() {
	s.rejected = true
}

// Helper method, get map based on State
func (s *Solution) byState(state ClusterNodeState) map[int]*WorkerNode {
	if state == Idle {
		return s.Idle
	} else if state == Active {
		return s.Active
	}
	log.Printf("Errore in cluster.go solution byState: (%d)\n", state)
	os.Exit(1)
	return nil
}

// addWorkerNode adds a WorkerNode to the Idle map, based on the assurance level
func (s *Solution) AddToSolution(state ClusterNodeState, wn *WorkerNode) {
	idleMap := s.Idle
	activeMap := s.Active
	_, exists_I := idleMap[wn.ID]
	_, exists_A := activeMap[wn.ID]
	if exists_I || exists_A {
		log.Printf("WorkerNode %d already exists in Active (%t) or Idle (%t)\n", wn.ID, exists_A, exists_I)
	} else {
		s.byState(state)[wn.ID] = wn
		s.n_replicas++
		// log.Printf("Added WorkerNode %d to Solution (State %s)\n", wn.ID, state)
	}
}

func (s Solution) String() string {
	result := fmt.Sprintf("Adding pod %d.\n", s.pod.ID)
	if s.rejected {
		result += "\tSolution not found:\n"
		return result
	}
	//else
	result += fmt.Sprintf("\t%d replicas deployed\n", s.n_replicas)
	result += "\tadd to Active Workers:\n"
	result += _state_String(s.Active)
	result += "\nwake Idle Workers:\n"
	result += _state_String(s.Idle)
	result += "\n"
	return result
}

func (s Solution) list_Ids() []int {
	ret := []int{}
	for id, _ := range s.Active {
		ret = append(ret, id)
	}
	for id, _ := range s.Idle {
		ret = append(ret, id)
	}
	return ret
}
