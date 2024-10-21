package main

import (
	"fmt"
	"os"
)

/* GLOBAL VARIABLES */
// Number of Worker Nodes
var n int = 21 //rand_ab_int(8, 25)

// Number of Pods
var m int = 414 //rand_ab_int(350, 700)

// Which algos am I comparing?
// var testNames []string = []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_leastAllocated"}
// var testNames []string = []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_mostAllocated"}
var testNames []string = []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_requestedToCapacityRatio"}

var nTests int = len(testNames)
var testCallables = []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{
	adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s,
}

// var scoringFunctions = []func(*WorkerNode, *Pod) float32{costAware_leastAllocated_score, costAware_leastAllocated_score, k8s_leastAllocated_score}
// var scoringFunctions = []func(*WorkerNode, *Pod) float32{costAware_mostAllocated_score, costAware_mostAllocated_score, k8s_mostAllocated_score}
var scoringFunctions = []func(*WorkerNode, *Pod) float32{costAware_requestedToCapacityRatio_score, costAware_requestedToCapacityRatio_score, k8s_requestedToCapacityRatio_score}


// Results (per test)
var Acceptance_Ratio [][]float32 = make([][]float32, m)
var Energy_cost_Ratio [][]float32 = make([][]float32, m)

// Worker Nodes replicas for each algorithm
var testClusters []*Cluster = make([]*Cluster, nTests)

var _MAX_ENERGY_COST = -1

const _Log Log = Log_None

// List of all Pods
var allPods []*Pod

//var allPods []*Pod = make([]*Pod, m)

/*	*	*	*	*	*	Initialization	*	*	*	*	*	*/
func init() {
	for t := range testNames {
		testClusters[t] = NewCluster(testNames[t])
	}

	/** Worker Nodes creation */
	for i := 0; i < n; i++ {
		wn := createRandomWorkerNode(i + 1 /*Id*/)
		fmt.Println(wn)
		// Every algo has the same nodes (replicas) inside
		for t := range testNames {
			testClusters[t].AddWorkerNode(wn.Copy())
		}
	}
	fmt.Printf("n: %d\tm: %d\n", n, m)
}

// MAIN
func main() {
	main_sequential()
}

func main_sequential() {
	var pod *Pod
	var cluster *Cluster

	// New Iteration (New Pod)
	for j := 0; j < m; j++ {
		/*Init row for results*/
		Acceptance_Ratio[j] = make([]float32, nTests+1)
		Energy_cost_Ratio[j] = make([]float32, nTests+1)

		Acceptance_Ratio[j][0] = float32(j)
		Energy_cost_Ratio[j][0] = float32(j)
		/*Adding new pod phase*/
		//Create Random Pod
		pod = createRandomPod(j)
		if _Log >= Log_Some {
			fmt.Println(pod)
		}
		// For each Cluster
		for t := range testNames {
			cluster = testClusters[t]
			var solution Solution
			solution = testCallables[t](cluster, pod, scoringFunctions[t])
			apply_solution(cluster, pod.Copy(), solution, testNames[t])

			//Results update
			Acceptance_Ratio[j][t+1] = (float32(cluster.accepted) / float32(cluster._Total_Pods))
			Energy_cost_Ratio[j][t+1] = (float32(cluster.energeticCost) / float32(cluster._Total_Energetic_Cost))

			if _Log >= Log_Some {
				fmt.Println(cluster)
				fmt.Println()
			}
			if _Log >= Log_Scores {
				fmt.Printf("\n%s:\t    \tAccepted: %d/%d (%.2f%%)\t\tEnergy: %d/%d (%.2f%%)\n\n",
					cluster.name,
					cluster.accepted, cluster._Total_Pods, 100.*Acceptance_Ratio[j][t+1],
					cluster.energeticCost, cluster._Total_Energetic_Cost, 100.*Energy_cost_Ratio[j][t+1],
				)
			}
		}
		/*Running pods*/
		for t := range testNames {
			cluster = testClusters[t]
			for _, wn := range cluster.All_list() {
				completed := wn.RunPods()
				if completed {
					cluster.DeactivateWorkerNode(wn.ID, wn.Assurance)
				}
			}
		}
		if _Log >= Log_Scores {
			fmt.Println()
		}
	}
	matrixToCsv("acceptance.csv", Acceptance_Ratio[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("energy.csv", Energy_cost_Ratio[:], append([]string{"pod index"}, testNames[:]...), 3)
}

func apply_solution(cluster *Cluster, pod *Pod, solution Solution, test_name string) {
	if solution.rejected {
		if _Log >= Log_Some {
			fmt.Printf("pod %d rejected by test %s\n", pod.ID, test_name)
		}
		cluster.PodRejected()
	} else {
		// Apply solution
		// Wake who needs to be awaken
		for _, assurance := range ASSURANCES {
			for id, wn := range solution.Idle.ByAssurance(assurance) {
				cluster.ActivateWorkerNode(id, assurance)
				wn.InsertPod(pod)
				// fmt.Println(wn)
			}
		}

		// and add to those already active
		for _, assurance := range ASSURANCES {
			for _, wn := range solution.Active.ByAssurance(assurance) {
				wn.InsertPod(pod)
				// fmt.Println(wn)
			}
		}

		cluster.PodAccepted()

		if _Log >= Log_All {
			// fmt.Printf("Solution applied by test %s\n", test_name)
			// fmt.Println(solution)
		}
	}
}

/** Greedy approach */
func adding_new_pod__greedy(cluster *Cluster, pod *Pod,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) Solution {
	// var replicas_left int = pod.replicas
	var required_replicas int = pod.replicas
	var solution Solution = NewSolution(pod)
	var exclude_ids Set = make(Set)

	var state_im_scanning ClusterNodeState = Active

	var no_eligible_highAssurance_node_left bool = false //set this to true if no High Assurance Idle is eligible. Use it to skip in final loop
	var no_eligible_lowAssurance_node_left bool = false  //same with low

	var id int = -1
	var score float32 = -1.
	var assurance Assurance = HighAssurance
	// Find best among Active nodes
	for required_replicas > 0 {

		if !no_eligible_highAssurance_node_left && (required_replicas >= 2 || no_eligible_lowAssurance_node_left) {
			// fmt.Printf("Searching in High, %s\n", state_im_scanning)
			// Search in High Assurance
			id, score = find_best_wn(cluster.byState(state_im_scanning).ByAssurance(HighAssurance), pod,
				true, exclude_ids,
				costAware_requestedToCapacityRatio_score, k8s_leastAllocated_condition,
			)
			if id == -1 {
				no_eligible_highAssurance_node_left = true
				continue
			}
			assurance = HighAssurance
			// fmt.Printf("Found wn %d, with score %.2f\n", id, score)

		} else if !no_eligible_lowAssurance_node_left && (required_replicas == 1 || no_eligible_highAssurance_node_left) {
			// fmt.Printf("Searching in Low, %s\n", state_im_scanning)
			// Search in Low Assurance
			id, score = find_best_wn(cluster.byState(state_im_scanning).ByAssurance(LowAssurance), pod,
				true, exclude_ids,
				costAware_requestedToCapacityRatio_score, k8s_leastAllocated_condition,
			)
			if id == -1 {
				no_eligible_lowAssurance_node_left = true
				continue
			}
			assurance = LowAssurance
			// fmt.Printf("Found wn %d, with score %.2f\n", id, score)

		} else { //No eligible Worker found
			if state_im_scanning == Active {
				// fmt.Println("No eligibile Active Worker left, searching for Idle ones")
				state_im_scanning = Idle

				no_eligible_highAssurance_node_left = false //set this to true if no High Assurance Idle is eligible. Use it to skip in final loop
				no_eligible_lowAssurance_node_left = false
				continue
			} else {
				// fmt.Println("No eligibile Idle Worker left, rejected pod ", pod.ID)
				solution.Reject()
				break
			}
		}
		// Found node
		best_node := cluster.byState(state_im_scanning).ByAssurance(assurance)[id]
		score = score

		exclude_ids.Add(id)
		solution.AddToSolution(state_im_scanning, best_node)
		required_replicas -= int(best_node.Assurance)
	}

	// fmt.Printf("%s\n", solution)
	return solution
}

func find_best_wn(nodes map[int]*WorkerNode, pod *Pod,
	//extra args
	check_eligibility bool, exclude_ids Set,
	placer_scoring_func func(*WorkerNode, *Pod) float32, placer_isBetter_eval_func func(float32, float32) bool,
) (int, float32) {

	var score float32
	var bestScore float32 = -1.
	var argbest int = -1
	var initialized bool = false
	for id, node := range nodes {
		// If node is not among excluded, AND is eligible (or you don't need to check for eligibility)
		if !exclude_ids.Contains(id) {
			if !check_eligibility || node.EligibleFor(pod) {
				// If bool is cheaper than checking arithmetically each time
				if !initialized {
					score = placer_scoring_func(node, pod)
					bestScore = score
					argbest = id
					initialized = true
				} else {
					score = placer_scoring_func(node, pod)
					if placer_isBetter_eval_func(score, bestScore) {
						bestScore = score
						argbest = id
					}
				}
				// fmt.Printf("Scoring WN %d: %f\n", id, score)
			}
			//else continue
		}
	}
	return argbest, bestScore
}

/* Dynamic */
/** Greedy approach */
func adding_new_pod__dynamic(cluster *Cluster, pod *Pod,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) Solution {

	var solution Solution = NewSolution(pod)

	matrix, references := _create_dynamic_programming_matrix(&cluster.Active, pod, pod.replicas, placer_scoring_func)
	missing := _update_solution(&solution, matrix, Active, references)

	if missing > 0 {
		matrix, references := _create_dynamic_programming_matrix(&cluster.Idle, pod, missing, placer_scoring_func)
		missing = _update_solution(&solution, matrix, Idle, references)
	}
	if missing > 0 {
		solution.Reject()
	}
	return solution
}

func _create_dynamic_programming_matrix(Nodes *ByAssurance, pod *Pod, R int,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) ([][]float32, []*WorkerNode) {
	/* Init vars*/
	var scores []float32
	var assurances []int
	var references []*WorkerNode
	var eligibles int = 0

	// Iterate over both maps
	for _, node := range Nodes.High {
		if node.EligibleFor(pod) {
			scores = append(scores, placer_scoring_func(node, pod))
			// fmt.Printf("Node %d. Score: %.2f\n", node.ID, placer_scoring_func(node, pod))
			assurances = append(assurances, int(node.Assurance))
			references = append(references, node)
			eligibles++
		}
	}

	for _, node := range Nodes.Low {
		if node.EligibleFor(pod) {
			scores = append(scores, placer_scoring_func(node, pod))
			// fmt.Printf("Node %d. Score: %.2f\n", node.ID, placer_scoring_func(node, pod))
			assurances = append(assurances, int(node.Assurance))
			references = append(references, node)
			eligibles++
		}
	}

	// Now, create a 2D matrix (array) with rows equal to the number of eligible nodes
	if len(scores) != len(assurances) || eligibles != len(scores) {
		fmt.Printf("Errore, scores (%d) e assurances (%d) hanno dimensione diversa (in dynamic)\n", len(scores), len(assurances))
		os.Exit(1)
	}

	N := eligibles              //len(scores)-1
	M := make([][]float32, N+1) // Create rows from 0 to N

	for i := range M {
		M[i] = make([]float32, R+1) // Create columns with size R
	}

	/*Init Matrix*/
	// Default: No Replicas needed
	for i := 0; i <= N; i++ {
		M[i][0] = 0.
	}
	// Default: No nodes left
	for r := 1; r <= R; r++ {
		M[0][r] = -1.
	}

	// Core Loop
	for i := 1; i <= N; i++ {
		vec_i := i - 1
		for r := 1; r <= R; r++ {
			preR := r - assurances[vec_i]
			if preR < 0 {
				preR = 0
			}

			// If no-pick is not enough
			if M[i-1][r] == -1. {
				//If also pick is not enough
				if M[i-1][preR] == -1. {
					M[i][r] = -1.
				} else {
					//Pick obligated
					M[i][r] = M[i-1][preR] + scores[vec_i]
				}
			} else {
				M[i][r] = min(M[i-1][r], M[i-1][preR]+scores[vec_i])
			}
		}
	}

	if false {
		fmt.Println()
		fmt.Printf("N\\R\t")
		for r := 0; r <= R; r++ {
			fmt.Printf("%d\t", r)
		}
		fmt.Println()
		for i := 0; i <= N; i++ {
			if i > 0 {
				fmt.Printf("%d(%d):\t", i, references[i-1].ID)
			} else {
				fmt.Printf("0(-):\t")
			}

			for r := 0; r <= R; r++ {
				fmt.Printf("%.2f\t", M[i][r])
			}
			fmt.Println()
		}
		fmt.Println()
	}
	return M, references
}

func _update_solution(solution *Solution, M [][]float32, state ClusterNodeState, references []*WorkerNode) int {
	N := len(M) - 1
	R := len(M[0]) - 1

	i := N
	r := R
	missing_replicas := 0
	for M[i][r] == -1 {
		r--
		missing_replicas++
	}
	for i > 0 && r > 0 {
		vec_i := i - 1
		if M[i][r] != M[i-1][r] {
			//Add i to solution
			// fmt.Printf("Create sol i:%d (ID: %d)\n", i, references[vec_i].ID)
			solution.AddToSolution(state, references[vec_i])
			r -= int(references[vec_i].Assurance)
		}
		i--
	}
	return missing_replicas
}

/*K8s*/
func adding_new_pod__k8s(cluster *Cluster, pod *Pod,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) Solution {
	// var replicas_left int = pod.replicas
	var required_replicas int = pod.replicas
	var solution Solution = NewSolution(pod)
	var exclude_ids Set = make(Set)

	var id int = -1
	// var score float32 = -1.
	// Find best among Active nodes
	var allNodes_map = cluster.All_map()
	for required_replicas > 0 {
		id, _ = find_best_wn(allNodes_map, pod, true, exclude_ids, placer_scoring_func, k8s_leastAllocated_condition)
		// fmt.Printf("Searching eligible for pod %d, got wn: %d \n", pod.ID, id)
		if id < 0 {
			solution.Reject()
			break
		}
		node, state, _ := cluster.Get_by_Id(id)
		solution.AddToSolution(state, node)
		exclude_ids.Add(id)
		required_replicas--
	}

	// fmt.Printf("%s\n", solution)
	return solution
}
