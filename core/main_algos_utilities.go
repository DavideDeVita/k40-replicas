package core

import (
	"fmt"
	"sort"

	cmn "local/KOrche/common"
)

// find_best_wn identifica il nodo con punteggio minimo tra un insieme di candidati.
//
// Parametri:
//   - nodes: mappa dei nodi candidati.
//   - pod: Pod da collocare.
//   - check_eligibility: se true, filtra i nodi non eleggibili.
//   - exclude_ids: set di nodi da ignorare (già usati o scartati).
//   - computed_scores: cache per evitare ricalcoli di score.
//   - placer_scoring_func: funzione di scoring usata per valutare i nodi.
//   - placer_isBetter_eval_func: funzione comparatrice (es. minore-is-better).
//
// Ritorna:
//   - id del nodo migliore e relativo punteggio.
//   - (-1, -1) se nessun nodo eleggibile trovato.
//
// Note:
//   - La funzione memorizza gli score già calcolati per ottimizzare cicli successivi.
//   - Esclude dinamicamente i nodi man mano che vengono giudicati non eleggibili.
func find_best_wn(nodes map[int]*WorkerNode,
	pod *Pod,
	//extra args
	check_eligibility bool,
	exclude_ids cmn.Set,
	computed_scores map[int]float32,
	scoringFunction func(*WorkerNode, *Pod) float32,
	placer_isBetter_eval_func func(float32, float32) bool,
) (int, float32) {
	/*This function is used by Greedy and K8s !*/
	var score float32
	var exists bool
	var bestScore float32 = -1.
	var argbest int = -1
	var initialized bool = false
	for id, node := range nodes {
		/* Salta nodi già visitati o scartati */
		if !exclude_ids.Contains(id) {
			if !check_eligibility || node.EligibleFor(pod) {
				// compute score and save it if not already computed
				score, exists = computed_scores[id]
				if !exists {
					score = scoringFunction(node, pod)
					/* Cache locale per evitare ricalcolo dello score */
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

			} else {
				// explain, reason := node.ExplainEligibility(pod)

				exclude_ids.Add(id) //Adding id to exclude it later
			}
		}
	}
	return argbest, bestScore
}

// sortByPrimary_Assurance ordina un vettore principale (assurances) e propaga lo stesso ordinamento
// a vettori secondari paralleli (scores, references, clusterstates).
//
// Parametri:
//   - assurances: vettore da ordinare (valore primario).
//   - secondari vari: vettori da riordinare coerentemente.
//   - cmp: funzione di confronto (true se a < b).
//   - ascending: se true, ordine crescente.
//
// Note:
//   - Mantenere la coerenza degli array è essenziale per la correttezza del DP.
//   - Tipicamente viene usata prima della costruzione della tabella DP.
func sortByPrimary_Assurance(primary_ass []float32,
	secondary_scores []float32,
	secondary_wn []*WorkerNode,
	secondary_cns []ClusterNodeState,
	condition func(a, b float32) bool,
	reverse bool) {
	n := len(primary_ass)

	// Create an index slice
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
	}

	// Sort indices based on primary array
	sort.Slice(indices, func(i, j int) bool {
		if !reverse {
			return condition(primary_ass[indices[i]], primary_ass[indices[j]])
		} else {
			return !condition(primary_ass[indices[i]], primary_ass[indices[j]])
		}
	})

	// Reorder primary array
	tempPrimary := make([]float32, n)
	tempSecondary := make([]float32, n)
	tempNodes := make([]*WorkerNode, n)
	tempStates := make([]ClusterNodeState, n)

	for i, idx := range indices {
		tempPrimary[i] = primary_ass[idx]
		tempSecondary[i] = secondary_scores[idx]
		tempNodes[i] = secondary_wn[idx]
		tempStates[i] = secondary_cns[idx]
	}

	copy(primary_ass, tempPrimary)
	copy(secondary_scores, tempSecondary)
	copy(secondary_wn, tempNodes)
	copy(secondary_cns, tempStates)
}

// DP_findEligibleSolution implementa la ricerca principale in Programmazione Dinamica.
//
// Parametri:
//   - n: numero di nodi considerati (già ordinati per assurance crescente).
//   - probabilities: vettore delle assurance dei nodi.
//   - theta: soglia di probabilità richiesta (es. criticality del pod).
//   - overkillSize: margine per accettare soluzioni più grandi della minima.
//
// Ritorna:
//   - dp: tabella 3D [i][j][k] con probabilità cumulative.
//   - firstEligible: soluzioni iniziali eleggibili.
//
// Note:
//   - Ogni cella dp[i][j][k] indica la probabilità di ottenere k successi
//     scegliendo j nodi tra i primi i.
//   - L’algoritmo si ferma alla prima soluzione per ogni j e tronca oltre overkillSize.
//   - È la parte computazionalmente più costosa di tutto l’algoritmo.
func DP_findEligibleSolution(n int,
	probabilities []float32,
	theta float32,
	maxAllowedReplicas int,
	overkillSize int,
) ([][][]float32, map[int][]int) {
	// Initialize a 3D DP table
	var maxJ int = maxAllowedReplicas
	if maxAllowedReplicas < 0 || maxAllowedReplicas >= n {
		maxJ = n
	}

	dp := make([][][]float32, n+1)
	for i := range dp {
		dp[i] = make([][]float32, maxJ+1) //n+1)
		for j := range dp[i] {
			dp[i][j] = make([]float32, maxJ+1) //n+1)
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
					var probIfInSolution float32 = 0.0
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

	// for _, fesol := range firstEligible{
	// 	fmt.Printf("[DEBUG] eligible solution of size %d: %v\n", len(fesol), fesol)
	// }

	return dp, firstEligible
}

// _DP_search_neigh_solutions genera tutte le soluzioni eleggibili a partire da un insieme iniziale.
// Combina due strategie:
//  1. _DP_permutateMinSolution → genera soluzioni vicine scambiando indici.
//  2. _DP_allGreaterTuples → esplora tuple con nodi a maggiore assurance.
//
// Ritorna un vettore di tuple di indici eleggibili.
func _DP_search_neigh_solutions(n int,
	probabilities []float32,
	theta float32,
	firstEligibles map[int][]int,
	neighSpan int,
	maxNeighbourgh int,
) [][]int {
	// Step 2: Generate all eligible solutions and their neighbors
	var allEligibles [][]int

	for _, elSol := range firstEligibles {
		// fmt.Printf("firstEligibles sol %v\n", elSol)

		// Get neighboring solutions
		eligiblesNeigh := _DP_permutateMinSolution(n, probabilities, theta, elSol, neighSpan, maxNeighbourgh)

		// fmt.Printf("%d neighbor solutions of size %d:\n", len(eligiblesNeigh), size)

		// Expand solutions
		allEligibles = append(allEligibles, _DP_allGreaterTuples(eligiblesNeigh, n)...)
	}

	return allEligibles
}

// _DP_permutateMinSolution esplora soluzioni “vicine” modificando gli indici di una soluzione minima.
//
// Parametri:
//   - n: numero totale di nodi.
//   - probabilities: assurance dei nodi.
//   - theta: soglia di eleggibilità.
//   - firstEligible: soluzione iniziale (tuple di indici).
//   - neighSearch: ampiezza della ricerca (quanto lontano cercare).
//
// Ritorna:
//   - Un insieme di tuple eleggibili generate per permutazione.
//
// Note:
//   - Il metodo incrementa e decrementa indici (swap 1-flip) e verifica l’eleggibilità.
//   - I duplicati sono eliminati per efficienza.
//   - È limitato da maxNeighborgh per evitare esplosione combinatoria.
func _DP_permutateMinSolution(n int,
	probabilities []float32,
	theta float32,
	firstEligible []int,
	neighSpan int,
	maxNeighborgh int,
) [][]int {
	eligiblesNeigh := [][]int{firstEligible}
	size := len(firstEligible)

	i := 0
	for (i < maxNeighborgh || maxNeighborgh < 0) && i < len(eligiblesNeigh) {
		_sol := eligiblesNeigh[i]
		if len(_sol) > 0 {
			for idx := 0; idx < size; idx++ { // Increment idx
				for jdx := 0; jdx < size; jdx++ { // Decrement jdx
					if idx == jdx {
						continue
					}

					for shift := 1; shift <= neighSpan; shift++ { // Shift values
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
						////// RECENTEMENTE INVERTITE PERCHè PENSO SIA PIù EFFICIENTE SE NON FUNZIONA REINVERTI
						if !cmn.ContainsSlice(eligiblesNeigh, neigh) {
							if _, ok := _DP_is_tuple_eligible(probabilities, theta, neigh); ok {
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

// _DP_allGreaterTuples genera in modo esaustivo tutte le tuple che includono
// almeno un nodo con assurance maggiore rispetto a quelle in input.
//
// Parametri:
//   - inputEligibles: soluzioni eleggibili di partenza.
//   - n: numero totale di nodi.
//
// Ritorna:
//   - Tutte le tuple valide (senza duplicati).
//
// Note:
//   - È una ricerca “brute-force” necessaria per garantire copertura completa delle soluzioni.
//   - Estremamente costosa: da usare con prudenza.
func _DP_allGreaterTuples(inputEligibles [][]int, n int) [][]int {
	eligibles := make(map[string][]int) // Use map to store unique slices as keys
	var results [][]int

	for _, el := range inputEligibles {
		eligibles[cmn.SliceToString(el)] = el

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
					key := cmn.SliceToString(newT)
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
		// if solp, ok := _DP_is_tuple_eligible(probabilities, theta, val); ok {
		// 	results = append(results, val)
		// } else {
		// 	fmt.Printf("[WARNING ERROR] Attempting to add solution %v to eligibles while its prob is %.3f\n", val, solp)
		// }
	}

	return results
}

// _DP_is_tuple_eligible verifica se una tupla di indici rappresenta una soluzione eleggibile.
// Converte gli indici in assurance, calcola la probabilità di avere almeno metà successi,
// e la confronta con la soglia theta.
//
// Ritorna:
//   - (probabilità, eleggibile)
func _DP_is_tuple_eligible(probabilities []float32,
	theta float32,
	tuple_sol []int,
) (float32, bool) {
	p_tuple := []float32{}
	for _, p_idx := range tuple_sol {
		p_tuple = append(p_tuple, probabilities[p_idx])
	}
	prob := cmn.Compute_probability_atLeastHalf(p_tuple)

	return prob, prob >= theta
}

// _DP_pick_top_solutions returns the top N eligible tuples with the lowest aggregated score.
// If outputsAmount >= number of eligibles, all tuples are returned sorted.
//
// Params:
//   - all_eligibles: list of candidate tuples
//   - all_nodes_scores: vector of per-node scores
//   - outputsAmount: number of top solutions to return
//   - aggregation_mode: "sum", "geometric", "squaredsum"
//
// Returns:
//   - solutions: list of tuples sorted by ascending score
//   - scores: corresponding aggregated scores, same order
func _DP_pick_top_solutions(
    all_eligibles [][]int,
    all_nodes_scores []float32,
    outputsAmount int,
    aggregation_mode string,
) ([][]int, []float32) {

    type scoredTuple struct {
        tuple []int
        score float32
    }

    // Compute all scores
    scored := make([]scoredTuple, 0, len(all_eligibles))

    for _, tuple := range all_eligibles {
        // Extract scores for nodes in tuple
        var scores []float32
        for _, node_idx := range tuple {
            scores = append(scores, all_nodes_scores[node_idx])
        }

        agg := cmn.AggregateScores(scores, aggregation_mode)
        fmt.Printf("[DEBUG] Score of %v is %.2f\n", tuple, agg)

        scored = append(scored, scoredTuple{tuple: tuple, score: agg})
    }

    // Sort by ascending score
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].score < scored[j].score
    })

    // Clamp N
    if outputsAmount <= 0 {
        outputsAmount = 1
    }
    if outputsAmount > len(scored) {
        outputsAmount = len(scored)
    }

    // Extract top N
    outTuples := make([][]int, outputsAmount)
    outScores := make([]float32, outputsAmount)

    for i := 0; i < outputsAmount; i++ {
        outTuples[i] = scored[i].tuple
        outScores[i] = scored[i].score
    }

    // fmt.Printf("[DEBUG] Top %d solutions extracted.\n", outputsAmount)
    return outTuples, outScores
}

// _update_solution costruisce la Solution finale a partire da una tupla di indici.
//
// Parametri:
//   - solution: puntatore alla Solution da aggiornare.
//   - selectedSolution: tupla di indici dei nodi scelti.
//   - references: vettore di puntatori ai nodi corrispondenti.
//   - probabilities: assurance associate ai nodi.
//   - scores: score dei singoli nodi.
//   - explainScoringFunc: funzione per generare il breakdown degli score.
//
// Effetti:
//   - Aggiunge i nodi selezionati alla soluzione finale e aggiorna il numero di repliche.
//   - Calcola e salva la probabilità cumulata per la soluzione complessiva.
func _DP_tuple_to_solution(	pod *Pod,
	tupleSolution []int,
	references []*WorkerNode,
	probabilities []float32,
	scores []float32,
	explainScoringFunc func(*WorkerNode, *Pod) map[string]float32,
) Solution {
	var solution Solution = InitSolution(pod)
	solutionProb, isIt := _DP_is_tuple_eligible(probabilities, pod.Criticality, tupleSolution)

	if !isIt {
		fmt.Printf("[ERROR] DP tuple solution not eligible\n\tProbH+ : %.3f / %.2f\n\tTuple: %v", solutionProb, solution.Pod.Criticality, tupleSolution)
	}

	// seen := make(map[int]bool)
	for _, nodeIdx := range tupleSolution {
		// Debugging: make sure no duplicates
		// if seen[nodeIdx] {
		// 	continue
		// }
		// seen[nodeIdx] = true

		node := references[nodeIdx]
		explainScore := explainScoringFunc(node, solution.Pod)
		solution.AddToSolution(node, solutionProb, scores[nodeIdx], explainScore)
	}
	return solution
}



