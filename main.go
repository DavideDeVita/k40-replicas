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
const n int = 20 //rand_ab_int(10, 25)

// Number of Pods
const m int = 5000 // rand_ab_int(5000, 10000)

const _MAX_REPLICAS int = 5
const _DP_Max_Neigh int = -1

const _DP_Xkeep = 1
const _DP_neighSpan = 2

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

const _Log LogLevel = Log_None
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
}

func parse_args() {
	_i_test := os.Args[1]
	switch _i_test {
	case "0": //custom
		_test = Test{
			name:           "custom_test",
			Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K4.0 Dynamic ALL", "K8s_mostAllocated"},
			Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__dynamic_allNodes, adding_new_pod__k8s},
			Is_multiparam:  []bool{true, true, true, false},

			Placing_scorer:  k8s_mostAllocated_score,
			Placing_w:       5,
			Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _computationPower_ratio, _log10_assurance_wasteless, _rt_waste},
			Multi_obj_w:     []float32{2, 2, 1, 1},
			Multi_obj_names: []string{"energy cost", "comput power", "log assurance", "rt waste"},
		}
		break
	case "1":
		_test = TEST_LeastAllocated
		break
	case "2":
		_test = TEST_MostAllocated
		break
	case "3":
		_test = TEST_RequestedToCapacityRatio
		break
	case "4":
		_test = TEST_LeastAllocated_5Params
		_test.name += "_4-2-2-1-1"
		break
	case "5":
		_test = TEST_MostAllocated_5Params
		_test.name += "_4-2-2-1-1"
		break
	case "6":
		_test = TEST_RequestedToCapacityRatio_5Params
		_test.name += "_4-2-2-1-1"
		break
	case "7":
		_test = TEST_LeastAllocated_5Params
		_test.Placing_w=5
		_test.Multi_obj_w= []float32{3.,2.,2.,1.}
		_test.name += "_5-3-2-2-1"
		// _test.name += fmt.Sprintf("_%d", int(_test.Placing_w))
		// for _, w := range(_test.Multi_obj_w){
		// 	_test.name += fmt.Sprintf("_%d", int(w))
		// }
		break
	case "8":
		_test = TEST_MostAllocated_5Params
		_test.Placing_w=5
		_test.Multi_obj_w= []float32{3.,2.,2.,1.}
		_test.name += "_5-3-2-2-1"
		break
	case "9":
		_test = TEST_RequestedToCapacityRatio_5Params
		_test.Placing_w=5
		_test.Multi_obj_w= []float32{3.,2.,2.,1.}
		_test.name += "_5-3-2-2-1"
		break
	case "10":
		_test = TEST_LeastAllocated_5Params
		_test.Placing_w=4
		_test.Multi_obj_w= []float32{2.,1.,1.,1.}
		_test.name += "_4-2-1-1-1"
		break
	case "11":
		_test = TEST_MostAllocated_5Params
		_test.Placing_w=4
		_test.Multi_obj_w= []float32{2.,1.,1.,1.}
		_test.name += "_4-2-1-1-1"
		break
	case "12":
		_test = TEST_RequestedToCapacityRatio_5Params
		_test.Placing_w=4
		_test.Multi_obj_w= []float32{2.,1.,1.,1.}
		_test.name += "_4-2-1-1-1"
		break
	case "13":
		_test = TEST_LeastAllocated_5Params
		_test.Placing_w=3
		_test.Multi_obj_w= []float32{2.,1.,1.,1.}
		_test.name += "_3-2-1-1-1"
		break
	case "14":
		_test = TEST_MostAllocated_5Params
		_test.Placing_w=3
		_test.Multi_obj_w= []float32{2.,1.,1.,1.}
		_test.name += "_3-2-1-1-1"
		break
	case "15":
		_test = TEST_RequestedToCapacityRatio_5Params
		_test.Placing_w=3
		_test.Multi_obj_w= []float32{2.,1.,1.,1.}
		_test.name += "_3-2-1-1-1"
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

			if _Log >= Log_Scores && t == nTests-1 && !const_array(R) {
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

		if _Log >= Log_Some {
			log.Println()
		}
	}
	matrixToCsv(_FOLDER+"acceptance.csv", Acceptance_Ratio[:], append([]string{"pod index"}, _test.Names[:]...), 3)
	matrixToCsv(_FOLDER+"energy.csv", Energy_cost_Ratio[:], append([]string{"pod index"}, _test.Names[:]...), 3)
	for t := range _test.Names {
		log.Printf("[%s] - completed in %s\n", _test.Names[t], readableNanoseconds(chronometers[t]))
		log.Println(testClusters[t])
	}
	log.Printf("n: %d\tm: %d\n", n, m)
	log.Printf("num rt: %d\n", _NumRT)
	log.Printf("Max Replicas: %d\n", _MAX_REPLICAS)
	log.Printf("DP Max Neigh: %d\n", _DP_Max_Neigh)
}

func apply_solution(cluster *Cluster, pod *Pod, solution Solution, test_name string) {
	if solution.rejected {
		if _Log >= Log_Some {
			log.Printf("[%s]\tpod %d rejected\n", test_name, pod.ID)
		}
		cluster.PodRejected()
	} else {
		if _Log >= Log_Some {
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
		if _MAX_REPLICAS > 0 && solution.n_replicas == _MAX_REPLICAS {
			solution.Reject()
			break
		}
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
			if _Log >= Log_Scores {
				log.Printf("[K4.0 Greedy]: With wn %d -> Prob h+: %.12f (theta = %.12f) \n", id, prob_atleast_half, theta)
			}
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
				if _Log >= Log_Scores {
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
	var edge_solutions map[int][]int

	// It is my responsability to create the vectors for assurances, scores and references
	var scores []float32
	var assurances []float64 // Probabilities
	var references []*WorkerNode
	var clusterstates []ClusterNodeState
	var n int
	var theta = pod.Criticality.value()

	if len(cluster.Active) > 0 {
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

			n = len(scores)
		}

		// Sort by Assurance asc
		sortByPrimary_f64(assurances, scores, references, clusterstates, func(a, b float64) bool {
			return a < b
		}, true)
		_, edge_solutions = DP_findEligibleSolution(n, assurances, theta, _DP_Xkeep)
	}

	// If no solution (that has theta or more) using only Active nodes, Try again on all
	if len(edge_solutions) == 0 {
		if true { //Writing the arrays of scores and assurance. If true just to fold it
			var overprice_func = func(node *WorkerNode) int {
				return int((node.EnergyCost * 15) / 10) // = *1.5
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

			n = len(scores)
		}

		// Sort by Assurance asc
		sortByPrimary_f64(assurances, scores, references, clusterstates, func(a, b float64) bool {
			return a < b
		}, true)
		_, edge_solutions = DP_findEligibleSolution(n, assurances, theta, _DP_Xkeep)

		/*If still no solution, reject*/
		if len(edge_solutions) == 0 {
			solution.Reject()
			return solution
		}
	}

	//One or more solution found
	all_eligibles := _DP_search_neigh_solutions(n, assurances, theta, edge_solutions, _DP_neighSpan)
	selectedSolution, _ := _DP_pick_best_solution(all_eligibles, scores)
	_update_solution(&solution, selectedSolution, references, clusterstates)
	return solution
}

func adding_new_pod__dynamic_allNodes(cluster *Cluster, pod *Pod,
	placer_scoring_func func(*WorkerNode, *Pod) float32,
) Solution {

	var solution Solution = NewSolution(pod)
	var edge_solutions map[int][]int

	// It is my responsability to create the vectors for assurances, scores and references
	var scores []float32
	var assurances []float64 // Probabilities
	var references []*WorkerNode
	var clusterstates []ClusterNodeState
	var n int
	var theta = pod.Criticality.value()

	if len(cluster.Active) > 0 {
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
		}
	}

	if true { //Writing the arrays of scores and assurance. If true just to fold it
		var overprice_func = func(node *WorkerNode) int {
			return int((node.EnergyCost * 3) / 2) // = *1.5
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
	}

	// Check for errors
	if len(scores) != len(assurances) {
		log.Printf("Errore, scores (%d) e assurances (%d) hanno dimensione diversa (in dynamic)\n", len(scores), len(assurances))
		os.Exit(1)
	}

	n = len(scores)
	// Sort by Assurance asc
	sortByPrimary_f64(assurances, scores, references, clusterstates, func(a, b float64) bool {
		return a < b
	}, true)
	_, edge_solutions = DP_findEligibleSolution(n, assurances, theta, _DP_Xkeep)

	/*If still no solution, reject*/
	if len(edge_solutions) == 0 {
		solution.Reject()
		return solution
	}

	//One or more solution found
	all_eligibles := _DP_search_neigh_solutions(n, assurances, theta, edge_solutions, _DP_neighSpan)
	selectedSolution, _ := _DP_pick_best_solution(all_eligibles, scores)
	_update_solution(&solution, selectedSolution, references, clusterstates)
	return solution
}

// findEligibleSolution implements the 3D DP approach in Go
func DP_findEligibleSolution(n int, probabilities []float64, theta float64, overkillSize int) ([][][]float64, map[int][]int) {
	// Initialize a 3D DP table
	var maxJ int = _MAX_REPLICAS
	if _MAX_REPLICAS < 0 || _MAX_REPLICAS>=n {
		maxJ = n
	}

	dp := make([][][]float64, n+1)
	for i := range dp {
		dp[i] = make([][]float64, maxJ+1) //n+1)
		for j := range dp[i] {
			dp[i][j] = make([]float64, maxJ+1) //n+1)
		}
	}
	dp[0][0][0] = 1.0 // Base case

	firstEligible := make(map[int][]int)
	firstEligibleSize := -1

	// Fill DP Table
	for i := 1; i <= n; i++ { // Nodes
		for j := 0; j <= i; j++ { // Selected Nodes
			if j > maxJ || (firstEligibleSize > 0 && j > firstEligibleSize+overkillSize) {
				continue
			}

			for k := 0; k <= j; k++ { // Successes
				if j > 0 {
					dp[i][j][k] += dp[i-1][j-1][k] * (1 - probabilities[i-1]) // Fails
					if k > 0 {
						dp[i][j][k] += dp[i-1][j-1][k-1] * probabilities[i-1] // Succeeds
					}
				} else {
					dp[i][j][k] = 1.0
				}
			}

			// Evaluate Solution
			if j > 0 {
				if firstEligibleSize < 1 || j < firstEligibleSize {
					probIfInSolution := 0.0
					for k := (j/2 + 1); k <= j; k++ {
						probIfInSolution += dp[i][j][k]
					}

					if probIfInSolution >= theta {
						// Backtrack
						sol := []int{i - 1}
						_i, _j := i-1, j-1
						for _j > 0 {
							if dp[_i][_j][_j] != dp[_i-1][_j][_j] {
								sol = append(sol, _i-1)
								_j--
							}
							_i--
						}

						firstEligible[j] = sol
						firstEligibleSize = j
					}
				}
			}
		}
	}

	// Remove extra solutions
	for size := firstEligibleSize + overkillSize + 1; size < maxJ; size++ { // size < n
		delete(firstEligible, size)
	}
	// log.Println("[Dynamic Debug] el solution: ", firstEligible)
	return dp, firstEligible
}

// Concats permutateMinSolution and AllGreaterTuples
func _DP_search_neigh_solutions(n int, probabilities []float64, theta float64, firstEligibles map[int][]int, neighSearch int) [][]int {
	// Step 2: Generate all eligible solutions and their neighbors
	var allEligibles [][]int

	for _, elSol := range firstEligibles {
		// fmt.Printf("Solution of size %d: %v\n", size, elSol)

		// Get neighboring solutions
		eligiblesNeigh := _DP_permutateMinSolution(n, probabilities, theta, elSol, neighSearch)

		// fmt.Printf("%d neighbor solutions of size %d:\n", len(eligiblesNeigh), size)

		// Expand solutions
		allEligibles = append(allEligibles, _DP_allGreaterTuples(eligiblesNeigh, n)...)
	}

	return allEligibles
}

func _DP_permutateMinSolution(n int, probabilities []float64, theta float64, firstEligible []int, neighSearch int) [][]int {
	eligiblesNeigh := [][]int{firstEligible}
	size := len(firstEligible)

	i := 0
	for (i < _DP_Max_Neigh || _DP_Max_Neigh<0) && i < len(eligiblesNeigh) {
		_sol := eligiblesNeigh[i]
		if len(_sol) > 0 {
			for idx := 0; idx < size; idx++ { // Increment idx
				for jdx := 0; jdx < size; jdx++ { // Decrement jdx
					if idx == jdx {
						continue
					}

					for shift := 1; shift <= neighSearch; shift++ { // Shift values
						neigh := append([]int(nil), _sol...) // Copy slice
						neigh[idx] += shift
						neigh[jdx] -= shift

						// Constraints to avoid duplicates or invalid sequences
						if neigh[idx] >= n || neigh[jdx] < 0 ||
							(idx > 0 && neigh[idx] >= _sol[idx-1]) ||
							(jdx < size-1 && neigh[jdx] <= _sol[jdx+1]) ||
							(idx > 0 && neigh[idx] >= neigh[idx-1]) ||
							(jdx < size-1 && neigh[jdx] <= neigh[jdx+1]) {
							// continue
							break //maybe optimization as early exit
						}

						// Check if eligible
						if _DP_is_tuple_eligible(probabilities, theta, neigh) {
							if !containsSlice(eligiblesNeigh, neigh) {
								eligiblesNeigh = append(eligiblesNeigh, neigh)
							}
						}
					}
				}
			}
		}
		i++
	}

	return eligiblesNeigh
}

func _DP_allGreaterTuples(inputEligibles [][]int, n int) [][]int {
	eligibles := make(map[string][]int) // Use map to store unique slices as keys
	var results [][]int

	for _, el := range inputEligibles {
		// log.Println("[Dynamic Debug] all greater tuples of: ", el)
		eligibles[sliceToString(el)] = el

		size := len(el)
		if true {
			newT := append([]int(nil), el...) // Copy slice

			for newT[0] < n {
				newT[size-1]++ // Increment last element

				// Adjust elements to maintain order
				if size > 1 && (newT[size-1] >= newT[size-2]) {
					idx := size - 1
					for idx > 0 && newT[idx] >= newT[idx-1] {
						newT[idx] = el[idx]
						newT[idx-1]++
						idx--
					}
				}

				// Ensure uniqueness
				if newT[0] < n {
					key := sliceToString(newT)
					if _, exists := eligibles[key]; !exists {
						eligibles[key] = append([]int(nil), newT...)

						// log.Println("\t\tfound: ", newT)
					}
				}
			}
		}
	}

	// Convert map values to slice
	for _, val := range eligibles {
		results = append(results, val)
	}

	return results
}

func _DP_is_tuple_eligible(probabilities []float64, theta float64, tuple_sol []int) bool {
	p_tuple := []float64{}
	for _, p_idx := range tuple_sol {
		p_tuple = append(p_tuple, probabilities[p_idx])
	}
	return compute_probability_atLeastHalf(p_tuple) >= theta
}

func _DP_pick_best_solution(all_eligibles [][]int, scores []float32) ([]int, float32) {
	var best_sol []int
	var min_score float32 = -1.
	for _, tuple := range all_eligibles {
		var score float32 = 0.
		for _, node_idx := range tuple {
			score += scores[node_idx]
		}

		if min_score < 0 || score < min_score {
			min_score = score
			best_sol = tuple
		}
		// log.Println("[Dynamic Debug] eligible: ", tuple, " with score: ", score)
	}
	return best_sol, min_score
}

func _update_solution(solution *Solution, selectedSolution []int, references []*WorkerNode, states []ClusterNodeState) int {
	for _, node_idx := range selectedSolution {
		if _Log >= Log_Scores {
			log.Printf("[Dynamic] best entry: node %d\n", references[node_idx].ID)
		}
		solution.AddToSolution(states[node_idx], references[node_idx])
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
		if _MAX_REPLICAS > 0 && solution.n_replicas == _MAX_REPLICAS {
			solution.Reject()
			break
		}
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
		if _Log >= Log_Scores {
			log.Printf("[K8s]: with wn %d -> Prob h+: %.12f (theta = %.12f) \n", id, prob_atleast_half, theta)
		}
	}

	// log.Printf("%s\n", solution)
	return solution
}
