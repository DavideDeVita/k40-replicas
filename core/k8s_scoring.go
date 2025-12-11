package core

import (
	cmn "local/KOrche/common"
	"math"
)

/********************** K8s Scorer **********************/

func K8sLeastAllocatedScore(node *WorkerNode, p *Pod) float32 {
	// Compute the percentage of allocated resources for Basic Resource
	allocCPU_p, allocMemory_p, allocStorage_p := node.RequestedPercentage()

	// Return the average of Basic Resources alloc percentages
	return (allocCPU_p + allocMemory_p + allocStorage_p) / 3.
}

func K8sMostAllocatedScore(node *WorkerNode, p *Pod) float32 {
	// Compute the percentage of freeated resources for Basic Resource
	freeCPU_p, freeMemory_p, freeStorage_p := node.UnrequestedPercentage()

	// Return the average of Basic Resources free percentages       NB: I am always searching the min
	return (freeCPU_p + freeMemory_p + freeStorage_p) / 3.
}

func K8sRequestedToCapacityRatioScore(node *WorkerNode, p *Pod) float32 {
	// Compute the requested-to-capacity ratio for each Basic Resource
	cpuRatio := float32(p.CPU.Requested+node.CPU.Requested) / float32(node.CPU.Capacity)
	memoryRatio := float32(p.Memory.Requested+node.Memory.Requested) / float32(node.Memory.Capacity)
	ramRatio := float32(p.Storage.Requested+node.Storage.Requested) / float32(node.Storage.Capacity)

	// var _cpu_weight, _Memory_weight, _ram_weight float32 = 1., 1., 1.
	// cpuRatio *= _cpu_weight
	// memoryRatio *= _Memory_weight
	// ramRatio *= _ram_weight

	// I invert it (1 - ...) because i am always looking for the minimum
	return 1. - ((cpuRatio + memoryRatio + ramRatio) / 3.)
}

/********************** Other params aware **********************/
/** I always want to minimize this:
  between Idle nodes: obviously
  between Active nodes: I want to add to the least costly, hoping the most costly will shutdown eventually
*/

func EnergyCostRatio_absolute(node *WorkerNode, pod *Pod) float32 {
	if MaxEnergyCost == 0 {
		return 0
	}
	return node.ComputeAbsoluteEnergyCost() / float32(MaxEnergyCost)
}

func EnergyCostRatio_linear(node *WorkerNode, pod *Pod) float32 {
	if MaxEnergyCost == 0 {
		return 0
	}
	return node.ComputeLinearEnergyCost(pod) / float32(MaxEnergyCost)
}

func DeltaEnergyCostRatio_absolute(node *WorkerNode, pod *Pod) float32 {
	if MaxClusterEnergyCost == 0 {
		return 0
	}
	if !node.IsActive() {
		return node.ComputeAbsoluteEnergyCost() / MaxClusterEnergyCost
	}
	return 0
}

func DeltaEnergyCostRatio_linear(node *WorkerNode, pod *Pod) float32 {
	if MaxClusterEnergyCost == 0 {
		return 0
	}
	delta := node.ComputeLinearEnergyCost(pod) - node.ComputeLinearEnergyCost(nil)
	return delta / MaxClusterEnergyCost
}

//Computation.. unused
// @Removed
// func _computationPower_ratio(node *WorkerNode, pod *Pod) float32 {
// 	// return 1. / float32(node.Computation_Power)
// 	return 2. / float32(node.Computation_Power+1)
// }

// Assurance
func SigmaNodeAssurance(node *WorkerNode, pod *Pod) float32 {
	return cmn.Sigma_f32(node.Assurance, 1.5)
}

// A differenza di SigmaNodeAssurance che considera solo l'assurance del nodo.. questa scala in base anche alla criticità del Pod
func SigmaAssuranceWasteless(node *WorkerNode, pod *Pod) float32 {
	if node.Assurance >= pod.Criticality {
		return 0.
	}

	// Rapporto tra l'Assurance e la Criticità:
	//	essendo A<C.. il rapporto è sempre in ]0..1[
	//	tende a 0.. quanto più l'assurance è più piccola della criticità (nodo poco adatto)
	//	tende a 1.. qunto più Assurance e Criticità si assomigliano (nodo più adatto)
	ac_ratio := node.Assurance / pod.Criticality

	//Sigma è una funzione monotona decrescente in [0..1]
	//	Per x = 0, Sigma = 1
	//	Per x = 1, Sigma = 0
	return cmn.Sigma_f32(ac_ratio, 1.5)
}

// RT
func RTWaste(node *WorkerNode, pod *Pod) float32 {
	if node.RealTime && !pod.RealTime {
		return 1.
	}
	return 0.
}

//	NEW
//
// Richiesta da ITALTEL, fa iniziare lo lo scoring dell'overcommit ad una percentuale della capacity
var ResourceDangerRatio float32 = 1.

// Limit Overcommit
// _limit_overcommit_penalty valuta quanto un Pod rischia di superare
// la capacità di un nodo in termini di limit (CPU, memoria, storage).
//
// Restituisce un punteggio in [0,1]:
//
//	0 → nessuno sforamento (perfetto)
//	1 → sforamento massimo (limite teoricamente infinito)
//
// La penalità aggrega CPU, memoria e storage come media delle tre penalità individuali.
func LimitOvercommitPenalty(node *WorkerNode, p *Pod) float32 {
	var overCommit float32 = 0.
	overCommit += _overcommit_EPiExcess(node.CPU.Capacity, node.CPU.Limit, p.CPU.Limit)
	overCommit += _overcommit_EPiExcess(node.Memory.Capacity, node.Memory.Limit, p.Memory.Limit)
	overCommit += _overcommit_EPiExcess(node.Storage.Capacity, node.Storage.Limit, p.Storage.Limit)

	return overCommit / 3.
}

// _overcommitRatio calcola la penalità normalizzata per una singola risorsa.
//
//	Non in uso.. sostituita da EPi_Excess
//
// Esempio:
//
//	capacity = 1000, totalLimit = 900  →  0.0  (nessuno sforamento)
//	capacity = 1000, totalLimit = 1200 →  0.2  (20% overcommit)
//	capacity = 1000, totalLimit = 3000 →  1.0  (penalità massima)
func _overcommitRatio(capacity, nodeLimit, podLimit int) float32 {
	totalLimit := float32(nodeLimit + podLimit)
	capF := float32(capacity) * ResourceDangerRatio

	if totalLimit <= capF {
		return 0.0
	}

	return 1. - (capF / totalLimit)
}

// _EPi_excess é una follia di mia invenzione per permettere allo score di scalare velocemente e assestarsi a 1.
func _overcommit_EPiExcess(capacity, nodeLimit, podLimit int) float32 {
	totalLimit := float32(nodeLimit + podLimit)
	capF := float32(capacity) * ResourceDangerRatio

	if totalLimit <= capF {
		return 0.0
	}

	// capF/totalLimit := sempre in ]0;1[, tende a zero quanto più totalLimit > capF
	// totalLimit/capF := di quanto il nuovo limit sfora la capacità? Sempre in ]0; inf[.
	// 		 Fa da esponente, moltiplicato per e*Pi, ad una base in [0;1] che diventa sempre più piccola..
	//			Quindi fa scalare il valore molto velocemente a 0
	//		Il tutto invertito con 1 - x, quindi tende a 1
	EPi_ret := 1 - cmn.Power_f32(capF/totalLimit, (math.E*math.Pi)*(totalLimit/capF))

	return EPi_ret
}

/********************** Eval conditions **********************/

// No need to have 3 diff functions here
func ScoreEvaluation_condition(score float32, best float32) bool {
	return score < best
}
