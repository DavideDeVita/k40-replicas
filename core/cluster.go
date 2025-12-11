package core

import (
	"log"
	"os"
)

type ClusterNodeState int

const (
	Idle   ClusterNodeState = 0
	Active ClusterNodeState = 1
)

/******************************* Cluster *******************************/
// Cluster struct with Active and Idle, both of type ByAssurance
type Cluster struct {
	name   string
	Active map[int]*WorkerNode
	Idle   map[int]*WorkerNode

	accepted              int
	energeticCost         int
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
		_Total_Energetic_Cost: 0,
	}
}

// Current and Max energy cost of the cluster (Used by k8s_scoring.Energy_cost functions)
var MaxEnergyCost int = 0
var CurrClusterEnergyCost, MaxClusterEnergyCost float32 = 0, 0

func (c *Cluster) ComputeEnergyConstants(hp AlgorithmHyperparams) {
	switch hp.EnergyCostMode {
	case "linear", "lineardelta", "deltalinear":
		for _, wn := range c.All_list() {
			CurrClusterEnergyCost += wn.ComputeLinearEnergyCost(nil)
			MaxClusterEnergyCost += float32(wn.EnergyCostCoef) // * 1
			if wn.EnergyCostCoef > MaxEnergyCost {
				MaxEnergyCost = wn.EnergyCostCoef
			}
		}

	default: // "absolute", "delta", "deltaabsolute", "absolutedelta"
		for _, wn := range c.Active {
			CurrClusterEnergyCost += float32(wn.EnergyCostCoef) // * 1
			MaxClusterEnergyCost += float32(wn.EnergyCostCoef)  // * 1
			if wn.EnergyCostCoef > MaxEnergyCost {
				MaxEnergyCost = wn.EnergyCostCoef
			}
		}
		for _, wn := range c.Idle {
			CurrClusterEnergyCost += float32(wn.EnergyCostCoef) // * 1
			MaxClusterEnergyCost += float32(wn.EnergyCostCoef)
			if wn.EnergyCostCoef > MaxEnergyCost {
				MaxEnergyCost = wn.EnergyCostCoef
			}
		}
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
		ret[wn.AutoID] = wn
	}
	return ret
}

// ID is intended as AutoID
func (c *Cluster) GetNodeByID(id int) (*WorkerNode, ClusterNodeState) {
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
