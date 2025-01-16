package main

//go run basicResourceType.go cluster.go enum.go k8s_scoring.go main.go pod.go set.go test.go utils.go workernode.go

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

/* GLOBAL VARIABLES */
// Number of Worker Nodes
const n int = 10 //rand_ab_int(10, 25)

// Number of Pods
const m int = 1000 // rand_ab_int(500, 1000)

// Which algos am I comparing?
var _test Test = TEST_LeastAllocated

// var _test Test = TEST_LeastAllocated_4Params
// var _test Test = TEST_MostAllocated
// var _test Test = TEST_MostAllocated_4Params
// var _test Test = TEST_RequestedToCapacityRatio
// var _test Test = TEST_RequestedToCapacityRatio_3Params

var Scorers []func(*WorkerNode, *Pod) float32
var chronometers []int64

var nTests int
var folderName string

// Results (per test)
var Acceptance_Ratio [][]float32 = make([][]float32, m)
var Energy_cost_Ratio [][]float32 = make([][]float32, m)

// Worker Nodes replicas for each algorithm
var testClusters []*Cluster

var _MAX_ENERGY_COST = -1

const DP_Xkeep = 1

const _Log LogLevel = Log_All
const _log_on_stdout bool = false

var logFile *os.File

// List of all Pods
var allPods []*Pod

//var allPods []*Pod = make([]*Pod, m)

/*	*	*	*	*	*	Initialization	*	*	*	*	*	*/
func init() {
	parse_args() // Gives value to _test
	init_log()

	/*Initialization of parameters that depend on _test*/
	folderName = _test.name
	nTests = len(_test.Names)

	init_scoring_params(_test)

	/*Creation of the clusters*/
	testClusters = make([]*Cluster, nTests)
	chronometers = make([]int64, nTests)
	for t := range _test.Names {
		testClusters[t] = NewCluster(_test.Names[t])
		if _test.Is_multiparam[t] {
			Scorers = append(Scorers, evaluate_score)
		} else {
			Scorers = append(Scorers, _test.Placing_scorer)
		}
	}

	/** Worker Nodes creation */
	for i := 0; i < n; i++ {
		wn := createRandomWorkerNode(i + 1 /*Id*/)

		log.Println(wn)
		// Every algo has the same nodes (copies of 'n' random generated nodes) inside
		for t := range _test.Names {
			testClusters[t].AddWorkerNode(wn.Copy())
		}
	}
	log.Printf("n: %d\tm: %d\n", n, m)
	log.Printf("num rt: %d\n", _NumRT)
}

func parse_args() {
	_i_test := os.Args[1]
	switch _i_test {
	case "0": //custom
		_test = Test{
			name:           "custom_test",
			Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_mostAllocated"},
			Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
			Is_multiparam:  []bool{true, true, false},

			Placing_scorer:  k8s_mostAllocated_score,
			Placing_w:       4,
			Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _computationPower_ratio, _log10_assurance},
			Multi_obj_w:     []float32{2, 1, 1},
		}
		break
	case "1":
		_test = TEST_LeastAllocated
		break
	case "2":
		_test = TEST_LeastAllocated_4Params
		break
	case "3":
		_test = TEST_MostAllocated
		break
	case "4":
		_test = TEST_MostAllocated_4Params
		break
	case "5":
		_test = TEST_RequestedToCapacityRatio
		break
	case "6":
		_test = TEST_RequestedToCapacityRatio_4Params
		break
	default:
		os.Exit(2)
	}
}

func init_log() {
	// Create or open a log file
	_FOLDER = _test.name + "\\" + os.Args[2] + "\\"
	var filename string = _FOLDER + "output.log"
	os.MkdirAll(_FOLDER, os.ModePerm)
	_logFile, err := os.Create(filename)
	if err != nil {
		log.Println("Error creating log file:", err)
		os.Exit(2)
	}
	logFile = _logFile
	// Create a multi-writer to write to both the file and the console
	var multiWriter io.Writer
	if _log_on_stdout {
		multiWriter = io.MultiWriter(os.Stdout, logFile)
	} else {
		multiWriter = io.MultiWriter(logFile)
	}
	// Set up the logger to use the multi-writer
	log.SetOutput(multiWriter)
}

// MAIN
func main() {
	main_sequential()
	logFile.Close()
}

func main_sequential() {
	var pod *Pod
	var cluster *Cluster
	var stopwatch time.Time
	var R []int = make([]int, nTests)

	// New Iteration (New Pod)
	for j := 0; j < m; j++ {
		/*Init row for storing results (these will be written in a csv at the end)*/
		Acceptance_Ratio[j] = make([]float32, nTests+1)
		Energy_cost_Ratio[j] = make([]float32, nTests+1)
		//first column is the index, unnecessary but i already wrote the plotting considering it
		Acceptance_Ratio[j][0] = float32(j)
		Energy_cost_Ratio[j][0] = float32(j)

		/*Adding new pod phase*/
		//Create Random Pod
		pod = createRandomPod(j)
		if _Log >= Log_Some {
			log.Println(pod)
		}
		// For each Cluster in the testbed, try to insert the pod (and its replicas)
		for t := range _test.Names {
			cluster = testClusters[t]
			var solution Solution
			//Start the chronometer
			stopwatch = time.Now()
			solution = _test.Algo_callables[t](cluster, pod, Scorers[t]) // Solution is an "insertion plan"
			apply_solution(cluster, pod.Copy(), solution, _test.Names[t])
			chronometers[t] += time.Since(stopwatch).Nanoseconds()
			//

			//Results update
			Acceptance_Ratio[j][t+1] = (float32(cluster.accepted) / float32(cluster._Total_Pods))
			Energy_cost_Ratio[j][t+1] = (float32(cluster.energeticCost) / float32(cluster._Total_Energetic_Cost))

			if _Log >= Log_Some {
				log.Println(cluster)
				log.Println()
			}
			if _Log == Log_Scores {
				log.Printf("\n%s:\t    \tAccepted: %d/%d (%.2f%%)\t\tEnergy: %d/%d (%.2f%%)\n\n",
					cluster.name,
					cluster.accepted, cluster._Total_Pods, 100.*Acceptance_Ratio[j][t+1],
					cluster.energeticCost, cluster._Total_Energetic_Cost, 100.*Energy_cost_Ratio[j][t+1],
				)
			}

			//replicas benchmark
			if solution.rejected {
				R[t] = 0
			} else {
				R[t] = solution.n_replicas
			}

			if t == nTests-1 && !const_array(R) {
				// log.Printf("[Replicas D]\n\t[%s]: \t%d\n\t[%s]: \t%d\n\t[%s]: \t%d\n\n", _test.Names[0], R[0], _test.Names[1], R[1], _test.Names[2], R[2])
				str := "[Replicas D]"
				for _i, _ := range R {
					str += fmt.Sprintf("\n\t[%s]: \t%d", _test.Names[_i], R[_i])
				}
				log.Printf("%s\n\n", str)
			}
		}

		/*Running pods; some may complete, nodes may be shut down and stuff*/
		for t, tag := range _test.Names {
			cluster = testClusters[t]
			for _, wn := range cluster.All_list() {
				completed := wn.RunPods(tag)
				if completed {
					cluster.DeactivateWorkerNode(wn.ID)
				}
			}
		}

		if _Log >= Log_Scores {
			log.Println()
		}
	}
	matrixToCsv(_FOLDER+"acceptance.csv", Acceptance_Ratio[:], append([]string{"pod index"}, _test.Names[:]...), 3)
	matrixToCsv(_FOLDER+"energy.csv", Energy_cost_Ratio[:], append([]string{"pod index"}, _test.Names[:]...), 3)
	for t := range _test.Names {
		log.Printf("[%s] - completed in %s\n", _test.Names[t], readableNanoseconds(chronometers[t]))
	}
}

func apply_solution(cluster *Cluster, pod *Pod, solution Solution, test_name string) {
	if solution.rejected {
		if _Log >= Log_Some {
			log.Printf("[%s]\tpod %d rejected\n", test_name, pod.ID)
		}
		cluster.PodRejected()
	} else {
		if _Log >= Log_Scores {
			log.Printf("[%s]\tSolution with %d replicas (%d)\n", test_name, solution.n_replicas, solution.list_Ids())
		}

		// Apply solution
		// Wake who needs to be awaken
		for id, wn := range solution.Idle {
			cluster.ActivateWorkerNode(id)
			wn.InsertPod(pod)
			// log.Println(wn)
		}

		// and add to those already active
		for _, wn := range solution.Active {
			wn.InsertPod(pod)
			// log.Println(wn)
		}

		cluster.PodAccepted()

		if _Log >= Log_All {
			// log.Printf("Solution applied by test %s\n", test_name)
			// log.Println(solution)
		}
	}
}

/*** These are the functions in the Callables vector ***/

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * 		Greedy		 * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
func adding_new_pod__greedy(cluster *Cluster, pod *Pod,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) Solution {
	var solution Solution = NewSolution(pod)
	var exclude_ids Set = make(Set)
	var computed_scores = map[int]float32{}

	var state_im_scanning = Active

	var id int = -1
	var score float32 = -1.

	var probabilities = []float64{}
	var prob_atleast_half float64 = 0.
	var theta = float64(pod.Criticality)

	for prob_atleast_half < theta {
		/* Search greedily the best node */
		id, score = find_best_wn(cluster.byState(state_im_scanning), pod, "Greedy",
			true, exclude_ids, computed_scores,
			placer_scoring_func, k8s_leastAllocated_condition,
		)

		if id == -1 {
			//Node not found
			if state_im_scanning == Active {
				state_im_scanning = Idle
				continue
			} else {
				solution.Reject()
				break
			}
		} else { //else innecessario
			// Found node
			if _Log >= Log_All {
				log.Printf("[Greedy]\tSearching eligible for pod %d, got wn: %d with score %.2f\n", pod.ID, id, score)
			}

			best_node := cluster.byState(state_im_scanning)[id]

			exclude_ids.Add(id) // This set is used to mark the nodes (id) i already scanned, so I won't scan over them again when I go from High to Low to High to Low again
			solution.AddToSolution(state_im_scanning, best_node)

			probabilities = append(probabilities, best_node.Assurance.value())
			prob_atleast_half = compute_probability_atLeastHalf(probabilities)
			log.Printf("[K4.0 Greedy]: Prob h+: %.12f (theta = %.12f) \n", prob_atleast_half, theta)
		}
	}

	// log.Printf("%s\n", solution)
	return solution
}

func find_best_wn(nodes map[int]*WorkerNode, pod *Pod,
	//extra args
	log_algo_name string,
	check_eligibility bool, exclude_ids Set, computed_scores map[int]float32,
	placer_scoring_func func(*WorkerNode, *Pod) float32, placer_isBetter_eval_func func(float32, float32) bool,
) (int, float32) {
	/*This function is used by Greedy and K8s !*/
	var score float32
	var exists bool
	var bestScore float32 = -1.
	var argbest int = -1
	var initialized bool = false
	for id, node := range nodes {
		// If node is not among excluded, AND is eligible (or you don't need to check for eligibility)
		// log.Printf("List of ecluded ids %s\n", exclude_ids)
		if !exclude_ids.Contains(id) {
			if !check_eligibility || node.EligibleFor(pod) {
				// compute score and save it if not already computed
				score, exists = computed_scores[id]
				if !exists {
					score = placer_scoring_func(node, pod)
					computed_scores[id] = score
				}

				// If bool is cheaper than checking arithmetically each time
				if !initialized {
					bestScore = score
					argbest = id
					initialized = true
				} else {
					if placer_isBetter_eval_func(score, bestScore) {
						bestScore = score
						argbest = id
					}
				}
				if _Log >= Log_Some {
					log.Printf("\tScoring wn %d: score %.2f\n", id, score)
				}
			} else {
				explain, reason := node.ExplainEligibility(pod)
				if explain && _Log >= Log_All {
					log.Printf("[%s]\tPod %d, Node %d ineligible. Reason %s\n", log_algo_name, pod.ID, node.ID, reason)
				}
				exclude_ids.Add(id) //Adding id to exclude it later
			}
		}
	}
	return argbest, bestScore
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * *      Dynamic        * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/* Dynamic */
/** Greedy approach */

func adding_new_pod__dynamic(cluster *Cluster, pod *Pod,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) Solution {

	var solution Solution = NewSolution(pod)

	// It is my responsability to create the vectors for assurances, scores and references
	var scores []float32
	var assurances []float64 // Probabilities
	var references []*WorkerNode
	var clusterstates []ClusterNodeState

	if true { //Writing the arrays of scores and assurance. If true just to fold it
		for _, node := range cluster.Active {
			if node.EligibleFor(pod) {
				// Filtering for eligibles
				scores = append(scores, placer_scoring_func(node, pod))
				assurances = append(assurances, node.Assurance.value())
				references = append(references, node)
				clusterstates = append(clusterstates, Active)

				if _Log >= Log_All {
					log.Printf("[Dynamic]\tNode %d, Pod %d. Score: %.2f\n", node.ID, pod.ID, scores[len(scores)-1])
				}
			} else {
				explain, reason := node.ExplainEligibility(pod)
				if explain && _Log >= Log_All {
					log.Printf("[Dynamic]\tNode %d, Pod %d. Reason %s\n", node.ID, pod.ID, reason)
				}
			}
		}

		// Check for errors
		if len(scores) != len(assurances) {
			log.Printf("Errore, scores (%d) e assurances (%d) hanno dimensione diversa (in dynamic)\n", len(scores), len(assurances))
			os.Exit(1)
		}
	}

	/*Run DP : byScores*/
	sortByPrimary_f32(scores, assurances, references, clusterstates, func(a, b float32) bool {
		return a < b
	}, true)
	_, el_solutions := _create_dynamic_programming_matrix(assurances, pod, DP_Xkeep)
	/*If no solution byScores, try byAssurance*/
	if len(el_solutions) == 0 {
		sortByPrimary_f64(assurances, scores, references, clusterstates, func(a, b float64) bool {
			return a < b
		}, true)
		_, el_solutions = _create_dynamic_programming_matrix(assurances, pod, DP_Xkeep)
	}

	// If no solution (that has theta or more) using only Active nodes, Try again on all
	if len(el_solutions) == 0 {
		if true { //Writing the arrays of scores and assurance. If true just to fold it
			var overprice_func = func(node *WorkerNode) int {
				return int((node.EnergyCost * 15) / 10) //power_floor(1.15, len(node.pods))
			}

			for _, node := range cluster.Idle {
				if node.EligibleFor(pod) {
					// Filtering for eligibles
					true_cost := node.EnergyCost
					node.EnergyCost = overprice_func(node)
					scores = append(scores, placer_scoring_func(node, pod))
					node.EnergyCost = true_cost

					assurances = append(assurances, node.Assurance.value())
					references = append(references, node)
					clusterstates = append(clusterstates, Idle)

					if _Log >= Log_All {
						log.Printf("[Dynamic]\tNode %d, Pod %d. Score: %.2f\n", node.ID, pod.ID, scores[len(scores)-1])
					}
				} else {
					explain, reason := node.ExplainEligibility(pod)
					if explain && _Log >= Log_All {
						log.Printf("[Dynamic]\tNode %d, Pod %d. Reason %s\n", node.ID, pod.ID, reason)
					}
				}
			}
			// Check for errors
			if len(scores) != len(assurances) {
				log.Printf("Errore, scores (%d) e assurances (%d) hanno dimensione diversa (in dynamic)\n", len(scores), len(assurances))
				os.Exit(1)
			}
		}

		/*Run DP : byScores*/
		sortByPrimary_f32(scores, assurances, references, clusterstates, func(a, b float32) bool {
			return a < b
		}, true)
		_, el_solutions := _create_dynamic_programming_matrix(assurances, pod, DP_Xkeep)
		/*If no solution byScores, try byAssurance*/
		if len(el_solutions) == 0 {
			sortByPrimary_f64(assurances, scores, references, clusterstates, func(a, b float64) bool {
				return a < b
			}, true)
			_, el_solutions = _create_dynamic_programming_matrix(assurances, pod, DP_Xkeep)
		}

		/*If still no solution, reject*/
		if len(el_solutions) == 0 {
			solution.Reject()
			return solution
		}
	}

	//One or more solution found
	var selectedSolution DP_Entry = choose_DP_entry(el_solutions, scores)
	_update_solution(&solution, selectedSolution, references, clusterstates)
	return solution
}

type DP_Entry struct {
	start             int
	end               int
	prob_atLeast_half float64
	score             float32
}

func _create_dynamic_programming_matrix(p []float64, pod *Pod,
	x_keep int,
) ([][][]float64, map[int][]DP_Entry) {
	// Init DP variables
	var first_el_amount int = -1
	eligibles := make(map[int][]DP_Entry)

	var theta float64 = pod.Criticality.value()
	n := len(p)

	// Create a 3D DP table initialized to 0.0
	dp := make([][][]float64, n)
	for i := 0; i < n; i++ {
		dp[i] = make([][]float64, n)
		for j := 0; j < n; j++ {
			dp[i][j] = make([]float64, n+1)
		}
	}

	// For each start node (i), up to each node (j), compute the probability that exactly k (for each k) satisfy their prob
	for i := 0; i < n; i++ { // Starting node
		for j := i; j < n; j++ { // Ending node
			this_amount := (j - i) + 1

			// Skip condition
			if first_el_amount != -1 && this_amount > first_el_amount+x_keep {
				continue
			}

			for k := 0; k <= this_amount; k++ { // Exact number of True variables
				if j == i { // Base case: single node in the range
					if k == 0 {
						dp[i][j][k] = 1 - p[j]
					} else if k == 1 {
						dp[i][j][k] = p[j]
					}
				} else { // General case: extend the range [i, j-1] to [i, j]
					dp[i][j][k] = dp[i][j-1][k] * (1 - p[j]) // j-th node is False
					if k > 0 {
						dp[i][j][k] += dp[i][j-1][k-1] * p[j] // j-th node is True
					}
				}
			}

			// Calculate probabilities for at least half the range
			// In go this should be put inside the previous loop to optimize, but like this is more clear
			prob_atleast_half := 0.0
			for k := (this_amount / 2) + 1; k <= this_amount; k++ {
				prob_atleast_half += dp[i][j][k]
			}

			if prob_atleast_half >= theta {
				if x_keep > 0 && first_el_amount == -1 {
					first_el_amount = this_amount
				}

				if _, exists := eligibles[this_amount]; !exists {
					eligibles[this_amount] = make([]DP_Entry, 0)
				}
				eligibles[this_amount] = append(eligibles[this_amount], DP_Entry{i, j, prob_atleast_half, -1})
			}
		}
	}

	return dp, eligibles
}

func choose_DP_entry(eligibles map[int][]DP_Entry, scores []float32) DP_Entry {
	var score float32 = 0.
	var best_score float32 = -1.
	var best_entry DP_Entry
	for _, entries := range eligibles {
		for _, entry := range entries {
			score = 0
			for i := entry.start; i <= entry.end; i++ {
				score += scores[i]
			}
			entry.score = score

			if best_score < 0 || score < best_score {
				best_score = score
				best_entry = entry
				// log.Println("New best entry", entry)
				// log.Printf("\t%d - %d : %.7f\t Score: %.4f\n\n", entry.start, entry.end, entry.prob_atLeast_half, score)
			}
		}
	}
	return best_entry
}

func _update_solution(solution *Solution, dp_entry DP_Entry, references []*WorkerNode, states []ClusterNodeState) int {
	for idx := dp_entry.start; idx <= dp_entry.end; idx++ {
		if _Log >= Log_Scores {
			log.Printf("[Dynamic] best entry: node %d\n", references[idx].ID)
		}
		solution.AddToSolution(states[idx], references[idx])
	}
	return solution.n_replicas
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * *		K8s		* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/*K8s*/
func adding_new_pod__k8s(cluster *Cluster, pod *Pod,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) Solution {
	var solution Solution = NewSolution(pod)
	var exclude_ids Set = make(Set)
	var computed_scores = map[int]float32{}

	var id int = -1
	var score float32 = -1.

	var probabilities = []float64{}
	var prob_atleast_half float64 = 0.
	var theta = float64(pod.Criticality)

	// Find best among ALL nodes
	var allNodes_map = cluster.All_map()

	for prob_atleast_half < theta {
		/* Search greedily the best node */
		id, score = find_best_wn(allNodes_map, pod, "K8s",
			true, exclude_ids, computed_scores,
			placer_scoring_func, k8s_leastAllocated_condition)

		if id < 0 {
			solution.Reject()
			break
		}
		//Found node
		if _Log >= Log_All {
			log.Printf("[K8s]\tSearching eligible for pod %d, got wn: %d with score %.2f\n", pod.ID, id, score)
		}

		node, state := cluster.Get_by_Id(id)

		exclude_ids.Add(id)
		solution.AddToSolution(state, node)

		probabilities = append(probabilities, node.Assurance.value())
		prob_atleast_half = compute_probability_atLeastHalf(probabilities)
		log.Printf("[K8s]: Prob h+: %.12f (theta = %.12f) \n", prob_atleast_half, theta)
	}

	// log.Printf("%s\n", solution)
	return solution
}
