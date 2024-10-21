package main

import (
	"fmt"
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

/******************************* ByAssurance *******************************/
type ByAssurance struct {
	High map[int]*WorkerNode
	Low  map[int]*WorkerNode
}

func NewByAssurance() ByAssurance {
	return ByAssurance{
		High: make(map[int]*WorkerNode),
		Low:  make(map[int]*WorkerNode),
	}
}

// Helper method, get map based on Assurance
func (ba ByAssurance) ByAssurance(assurance Assurance) map[int]*WorkerNode {
	if assurance == LowAssurance {
		return ba.Low
	} else if assurance == HighAssurance {
		return ba.High
	}
	fmt.Printf("Errore in cluster.go ByAssurance: (%d)\n", assurance)
	os.Exit(1)
	return map[int]*WorkerNode{}
}

func (ba ByAssurance) All_list() []*WorkerNode {
	ret := make([]*WorkerNode, 0)
	for _, wn := range ba.High {
		ret = append(ret, wn)
	}
	for _, wn := range ba.Low {
		ret = append(ret, wn)
	}
	return ret
}

func (ba ByAssurance) String() string {
	result := ""
	result += "\t\tHigh Assurance:"
	for id, node := range ba.High {
		result += fmt.Sprintf(" %d (%d),", id, len(node.pods))
	}
	result += "\n\t\tLow Assurance:"
	for id, node := range ba.Low {
		result += fmt.Sprintf(" %d (%d),", id, len(node.pods))
	}
	return result
}

func (ba ByAssurance) Size() int {
	return len(ba.High) + len(ba.Low)
}

/******************************* Cluster *******************************/
// Cluster struct with Active and Idle, both of type ByAssurance
type Cluster struct {
	name   string
	Active ByAssurance
	Idle   ByAssurance

	accepted              int
	energeticCost         int
	_Total_Pods           int
	_Total_Energetic_Cost int
}

// NewCluster initializes a Cluster with empty Active and Idle maps for both High and Low
func NewCluster(name string) *Cluster {
	return &Cluster{
		name:   name,
		Active: NewByAssurance(),
		Idle:   NewByAssurance(),

		accepted:              0,
		energeticCost:         0,
		_Total_Pods:           0,
		_Total_Energetic_Cost: 0,
	}
}

// Helper method, get map based on Assurance
func (c *Cluster) byState(state ClusterNodeState) ByAssurance {
	if state == Idle {
		return c.Idle
	} else if state == Active {
		return c.Active
	}
	fmt.Printf("Errore in cluster.go byState: (%d)\n", state)
	os.Exit(1)
	return ByAssurance{}
}

func (c Cluster) String() string {
	result := fmt.Sprintf("%s\n", c.name)
	result += fmt.Sprintf("Pod accepted: %d/%d (%.2f%%)\n", c.accepted, c._Total_Pods, 100.*float32(c.accepted)/float32(c._Total_Pods))
	result += fmt.Sprintf("Energetic cost: %d/%d (%.2f%%)\n", c.energeticCost, c._Total_Energetic_Cost, 100.*float32(c.energeticCost)/float32(c._Total_Energetic_Cost))
	//else
	result += fmt.Sprintf("\tActive Workers: %d\n", c.Active.Size())
	result += c.Active.String() + "\n"
	result += fmt.Sprintf("\tIdle Workers: %d\n", c.Idle.Size())
	result += c.Idle.String()
	return result
}

// addWorkerNode adds a WorkerNode to the Idle map, based on the assurance level
func (c *Cluster) AddWorkerNode(wn *WorkerNode) {
	var assurance Assurance = wn.Assurance
	idleMap := c.Idle.ByAssurance(assurance)
	activeMap := c.Active.ByAssurance(assurance)

	_, exists_I := idleMap[wn.ID]
	_, exists_A := activeMap[wn.ID]
	if !exists_I && !exists_A {
		idleMap[wn.ID] = wn
		c._Total_Energetic_Cost += wn.EnergyCost
		// fmt.Printf("Added WorkerNode %d to Idle (Assurance: %s)\n\n", wn.ID, assurance)
	} else {
		fmt.Printf("WorkerNode %d already exists in Active (%t) or Idle (%t) (Assurance: %s)\n", wn.ID, exists_A, exists_I, assurance)
		os.Exit(1)
	}
}

func (c *Cluster) All_list() []*WorkerNode {
	var ret []*WorkerNode = make([]*WorkerNode, 0)
	ret = append(ret, c.Active.All_list()...)
	ret = append(ret, c.Idle.All_list()...)
	return ret
}

func (c *Cluster) All_map() map[int]*WorkerNode {
	var ret map[int]*WorkerNode = make(map[int]*WorkerNode)
	for _, wn := range c.All_list() {
		ret[wn.ID] = wn
	}
	return ret
}

func (c *Cluster) Get_by_Id(id int) (*WorkerNode, ClusterNodeState, Assurance) {
	var node *WorkerNode
	var exists bool

	node, exists = c.Active.High[id]
	if exists {
		return node, Active, HighAssurance
	}

	node, exists = c.Active.Low[id]
	if exists {
		return node, Active, LowAssurance
	}

	node, exists = c.Idle.High[id]
	if exists {
		return node, Idle, HighAssurance
	}

	node, exists = c.Idle.Low[id]
	if exists {
		return node, Idle, LowAssurance
	}
	fmt.Println("Errore in get By Id", id)
	return nil, 0, 0
}

// activateWorkerNode moves a WorkerNode from Idle to Active, based on the assurance level
func (c *Cluster) ActivateWorkerNode(id int, assurance Assurance) {
	idleMap := c.Idle.ByAssurance(assurance)
	activeMap := c.Active.ByAssurance(assurance)

	if wn, idleExists := idleMap[id]; idleExists {
		if _, activeExists := activeMap[id]; !activeExists {
			delete(idleMap, id)
			activeMap[id] = wn
			c.energeticCost += wn.EnergyCost
			// fmt.Printf("Updated en cost for cluster %s: %d\n", c.name, c.energeticCost)
			if _Log>=Log_All{
				fmt.Printf("WorkerNode %d moved from Idle to Active (Assurance: %s)\n", id, assurance)
			}
		} else {
			fmt.Printf("WorkerNode %d is already Active (Assurance: %s)\n", id, assurance)
		}
	} else {
		fmt.Printf("WorkerNode %d not found in Idle (Assurance: %s)\n", id, assurance)
	}
}

// deactivateWorkerNode moves a WorkerNode from Active to Idle, based on the assurance level
func (c *Cluster) DeactivateWorkerNode(id int, assurance Assurance) {
	activeMap := c.Active.ByAssurance(assurance)
	idleMap := c.Idle.ByAssurance(assurance)

	if wn, idleExists := activeMap[id]; idleExists {
		if _, idleExists := idleMap[id]; !idleExists {
			delete(activeMap, id)
			idleMap[id] = wn
			c.energeticCost -= wn.EnergyCost
			if _Log>=Log_Some{
				fmt.Printf("WorkerNode %d moved from Active to Idle (Assurance: %s)\n", id, assurance)
			}
			// fmt.Printf("Updated en cost for cluster %s: %d\n", c.name, c.energeticCost)
		} else {
			fmt.Printf("WorkerNode %d is already Idle (Assurance: %s)\n", id, assurance)
		}
	} else {
		fmt.Printf("WorkerNode %d not found in Active (Assurance: %s)\n", id, assurance)
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
	pod      *Pod
	rejected bool
	Active   ByAssurance
	Idle     ByAssurance
}

// NewSolution initializes a Solution with empty Active and Idle maps for both High and Low
func NewSolution(pod *Pod) Solution {
	return Solution{
		pod:      pod,
		rejected: false,
		Active:   NewByAssurance(),
		Idle:     NewByAssurance(),
	}
}

func (s *Solution) Reject() {
	s.rejected = true
}

// Helper method, get map based on Assurance
func (s *Solution) byState(state ClusterNodeState) ByAssurance {
	if state == Idle {
		return s.Idle
	} else if state == Active {
		return s.Active
	}
	fmt.Printf("Errore in cluster.go solution byState: (%d)\n", state)
	os.Exit(1)
	return ByAssurance{}
}

// addWorkerNode adds a WorkerNode to the Idle map, based on the assurance level
func (s *Solution) AddToSolution(state ClusterNodeState, wn *WorkerNode) {
	var assurance Assurance = wn.Assurance
	idleMap := s.Idle.ByAssurance(assurance)
	activeMap := s.Active.ByAssurance(assurance)

	_, exists_I := idleMap[wn.ID]
	_, exists_A := activeMap[wn.ID]
	if exists_I || exists_A {
		fmt.Printf("WorkerNode %d already exists in Active (%t) or Idle (%t) (Assurance: %s)\n", wn.ID, exists_A, exists_I, assurance)
	} else {
		s.byState(state).ByAssurance(wn.Assurance)[wn.ID] = wn
		// fmt.Printf("Added WorkerNode %d to Solution (State %s) (Assurance: %s)\n", wn.ID, state, assurance)
	}
}

func (s Solution) String() string {
	result := fmt.Sprintf("Adding pod %d.\n", s.pod.ID)
	if s.rejected {
		result += "   Solution not found:\n"
		return result
	}
	//else
	result += "\tadd to Active Workers:\n"
	result += s.Active.String()
	result += "\nwake Idle Workers:\n"
	result += s.Idle.String()
	result += "\n"
	return result
}
