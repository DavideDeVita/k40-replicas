package api

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	core "local/KOrche/core"
)

//							 INPUT

// buildWorkerNodeFromDTO converts the WorkerNodeDTO to a core.WorkerNode.
func BuildWorkerNodeFromDTO(wn WorkerNodeDTO) core.WorkerNode {
	node := core.WorkerNode{
		ID:     wn.ID,
		AutoID: core.NextAutoID(),
		CPU: core.ResourceStats{
			Capacity:  parseResourceQty(wn.CPU.Capacity, "cpu"),
			Requested: parseResourceQty(wn.CPU.Requested, "cpu"),
			Limit:     parseResourceQty(wn.CPU.Limit, "cpu"),
		},
		Memory: core.ResourceStats{
			Capacity:  parseResourceQty(wn.Memory.Capacity, "memory"),
			Requested: parseResourceQty(wn.Memory.Requested, "memory"),
			Limit:     parseResourceQty(wn.Memory.Limit, "memory"),
		},
		Storage: core.ResourceStats{
			Capacity:  parseResourceQty(wn.Storage.Capacity, "storage"),
			Requested: parseResourceQty(wn.Storage.Requested, "storage"),
			Limit:     parseResourceQty(wn.Storage.Limit, "storage"),
		},
		RealTime:       wn.RealTime,
		EnergyCostCoef: wn.EnergyCostCoef,
		EnergyCostIdle: wn.EnergyCostIdle,
		Assurance:      wn.Assurance,
	}

	node.CPU.ComputeDerived()
	node.Memory.ComputeDerived()
	node.Storage.ComputeDerived()

	return node
}

// buildPodFromDTO converts a PodDTO to a core.Pod.
func BuildPodFromDTO(p PodDTO) core.Pod {
	return core.Pod{
		ID:          p.ID,
		RealTime:    p.RealTime,
		Criticality: p.Criticality,
		CPU: core.PodResources{
			Requested: parseResourceQty(p.CPU.Requested, "cpu"),
			Limit:     parseResourceQty(p.CPU.Limit, "cpu"),
		},
		Memory: core.PodResources{
			Requested: parseResourceQty(p.Memory.Requested, "memory"),
			Limit:     parseResourceQty(p.Memory.Limit, "memory"),
		},
		Storage: core.PodResources{
			Requested: parseResourceQty(p.Storage.Requested, "storage"),
			Limit:     parseResourceQty(p.Storage.Limit, "storage"),
		},
	}
}

//							 OUTPUT

// build Response from solution
func BuildResponseFromSolutions(pid string, sols []core.Solution) PlacementResult {
	resp := PlacementResult{
		PodID:       pid,
		Accepted:    false,
		Solutions: []SolutionDTO{},
	}

	for _, s := range sols {
		if !resp.Accepted && s.Accepted{
			resp.Accepted = true
		}

		resp.Solutions = append(resp.Solutions, BuildSDTOFromSolution(s))
	}

	return resp
}

func BuildSDTOFromSolution(sol core.Solution) SolutionDTO {
	resp := SolutionDTO{
		Replicas:    sol.Replicas,
		Probability: sol.Probability,
		Explanation: sol.Explanation,
		DeltaEnergy: sol.DeltaEnergy,
	}

	for _, n := range sol.Nodes {
		resp.Nodes = append(resp.Nodes, PlacementNodeDTO{
			ID:        n.NodeID,
			CPU:       BuildRSDTO_FromRS(n.CPU, "cpu"),
			Memory:    BuildRSDTO_FromRS(n.Memory, "memory"),
			Storage:   BuildRSDTO_FromRS(n.Storage, "storage"),
			Score:     n.Score,
			Subscores: n.ExplainScores,
		})
	}

	return resp
}

func BuildRSDTO_FromRS(rs core.ResourceStats, resourceType string) ResourceStatsDTO {
	return ResourceStatsDTO{
		Capacity:  formatResourceQty(rs.Capacity, resourceType),
		Requested: formatResourceQty(rs.Requested, resourceType),
		Limit:     formatResourceQty(rs.Limit, resourceType),
	}
}

// 						PARSE & FORMAT Resources

// parseResourceQty converts a Kubernetes-style resource quantity string into an integer base unit.
// - For CPU: returns millicores (m)
// - For Memory/Storage: returns bytes
// Accepts both SI (k, M, G, T, P, E) and IEC (Ki, Mi, Gi, Ti, Pi, Ei) suffixes.
func parseResourceQty(s string, resourceType string) int {
	if s == "" {
		return 0
	}

	s = strings.TrimSpace(s)
	re := regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?)\s*([a-zA-Z]*)$`)
	m := re.FindStringSubmatch(s)
	if len(m) != 3 {
		log.Printf("[WARN] cannot parse resource quantity: '%s'", s)
		return 0
	}

	num, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		log.Printf("[WARN] invalid numeric format in '%s': %v", s, err)
		return 0
	}

	unit := strings.ToLower(m[2])
	switch resourceType {
	case "cpu":
		// CPU: default = core, supports m (millicores)
		switch unit {
		case "", "core", "cores":
			return int(num * 1000)
		case "m":
			return int(num)
		default:
			log.Printf("[WARN] unknown CPU unit '%s', assuming millicores", unit)
			return int(num)
		}

	default:
		// Memory/Storage: base unit = bytes
		mult := float64(1)
		switch unit {
		case "k":
			mult = 1e3
		case "m":
			mult = 1e6
		case "g":
			mult = 1e9
		case "t":
			mult = 1e12
		case "p":
			mult = 1e15
		case "e":
			mult = 1e18
		case "ki":
			mult = math.Pow(2, 10)
		case "mi":
			mult = math.Pow(2, 20)
		case "gi":
			mult = math.Pow(2, 30)
		case "ti":
			mult = math.Pow(2, 40)
		case "pi":
			mult = math.Pow(2, 50)
		case "ei":
			mult = math.Pow(2, 60)
		case "":
			mult = 1 // assume bytes
		default:
			log.Printf("[WARN] unknown memory/storage unit '%s', assuming bytes", unit)
		}
		return int(num * mult)
	}
}

// formatResourceQty converts an integer back to a readable K8s-style quantity string.
// - For CPU: always returns millicores ("m")
// - For Memory/Storage: picks the largest binary unit without remainder (Ki, Mi, Gi, Ti)
func formatResourceQty(value int, resourceType string) string {
	if value == 0 {
		return "0"
	}

	switch resourceType {
	case "cpu":
		return strconv.Itoa(value) + "m"

	default:
		units := []struct {
			suffix string
			size   float64
		}{
			{"Ei", math.Pow(2, 60)},
			{"E", 1e18},
			{"Pi", math.Pow(2, 50)},
			{"P", 1e15},
			{"Ti", math.Pow(2, 40)},
			{"T", 1e12},
			{"Gi", math.Pow(2, 30)},
			{"G", 1e9},
			{"Mi", math.Pow(2, 20)},
			{"M", 1e6},
			{"Ki", math.Pow(2, 10)},
			{"K", 1e3},
		}

		v := float64(value)
		for _, u := range units {
			if v >= u.size && math.Mod(v, u.size) == 0 {
				return strconv.FormatFloat(v/u.size, 'f', -1, 64) + u.suffix
			}
		}

		// fallback → bytes
		return strconv.Itoa(value)
	}
}
