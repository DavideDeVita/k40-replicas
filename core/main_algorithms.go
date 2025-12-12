package core

import (
	"fmt"
	"log"
	"os"

	cmn "local/KOrche/common"
)

/*
Package core – Placement Algorithms
-----------------------------------
Questo file contiene le implementazioni principali degli algoritmi di placement.
Ogni variante implementa la stessa firma in modo da essere intercambiabile.

Algoritmi inclusi:
  - Greedy             → scelta iterativa del miglior nodo disponibile
  - DP_stateAware      → programmazione dinamica su nodi attivi, fallback sugli idle
  - DP_stateAgnostic   → DP su tutti i nodi contemporaneamente
  - K8s                → algoritmo “baseline” ispirato al leastAllocated di Kubernetes

Tutti restituiscono un oggetto Solution popolato o rifiutato.
*/

// AddingNewPodGreedy implementa la versione “greedy” del posizionamento di un Pod.
//
// Logica generale:
//  1. Inizializza una soluzione vuota.
//  2. Finché la probabilità cumulata di successo (prob_atLeastHalf) non supera la criticality richiesta dal Pod,
//     cerca il nodo con punteggio minimo usando find_best_wn().
//  3. Inserisce il nodo nella soluzione e aggiorna la probabilità complessiva.
//  4. Se nessun nodo eleggibile resta, oppure viene superato il numero massimo di repliche, rifiuta la soluzione.
//
// Differenze con K8s:
//   - Considera prima i nodi *Active* (già attivi) per minimizzare il costo energetico.
//   - Solo in seguito, se necessario, passa agli *Idle*.
//
// Parametri:
//   - cluster: stato attuale del sistema (nodi attivi e inattivi).
//   - pod: il Pod da posizionare.
//   - hyperparams.MaxReplicas: limite superiore per il numero di repliche (-1 = illimitato).
//   - scoringFunction: funzione di scoring globale.
//   - explainScoringFunc: versione estesa della funzione di scoring che restituisce un breakdown per sottofunzioni.
//
// Ritorna:
//   - Una Solution popolata o marcata come rifiutata.
func AddingNewPodGreedy(cluster *Cluster,
	pod *Pod,
	scoringFunction func(*WorkerNode, *Pod) float32,
	explainScoringFunc func(*WorkerNode, *Pod) map[string]float32,
	outputsAmount int,
	hyperparams AlgorithmHyperparams,
) []Solution {
	var solution Solution = InitSolution(pod)
	var exclude_ids cmn.Set = make(cmn.Set)
	var computed_scores = map[int]float32{}

	var state_im_scanning = Active

	var id int = -1

	var probabilities = []float32{}
	var prob_atleast_half float32 = -1.
	var theta = float32(pod.Criticality)
	var score float32
	var explainScore map[string]float32

	/* /Loop di selezione: continua finché non soddisfa la criticality */
	for prob_atleast_half < theta {
		if hyperparams.MaxReplicas > 0 && solution.Replicas == hyperparams.MaxReplicas {
			solution.Reject(fmt.Sprintf("Exceeded number of replicas (%d), while criticality constraint yet to meet (%.3f / %.3f)", hyperparams.MaxReplicas, prob_atleast_half, theta))
			break
		}
		/* Selezione del miglior nodo eleggibile */
		id, score = find_best_wn(cluster.byState(state_im_scanning), pod,
			true, exclude_ids, computed_scores,
			scoringFunction, ScoreEvaluation_condition,
		)

		if id == -1 {
			/* Nessun nodo trovato. Passa a Idle o rifiuta */
			if state_im_scanning == Active {
				state_im_scanning = Idle
				continue
			} else {
				solution.Reject(fmt.Sprintf("No more nodes to evaluate, while criticality constraint yet to meet (%.3f / %.3f)", prob_atleast_half, theta))
				break
			}

		} else { //else innecessario
			// Trovato un nodo
			best_node := cluster.byState(state_im_scanning)[id]

			/* Aggiunta dell'Assurance del nuovo nodo alla lista e calcolo del criterio di Criticità */
			probabilities = append(probabilities, best_node.Assurance)
			prob_atleast_half = cmn.Compute_probability_atLeastHalf(probabilities)

			exclude_ids.Add(id) // This set is used to mark the nodes (id) i already scanned, so I won't scan over them again

			explainScore = explainScoringFunc(best_node, pod)

			solution.AddToSolution(best_node, prob_atleast_half, score, explainScore)
		}
	}

	// Wrap it up
	if solution.Accepted {
		solution.WrapUpSolution(cluster, hyperparams)
	}

	return []Solution{solution}
}

// AddingNewPodDPStateAware implementa l’approccio di Programmazione Dinamica considerando prima i nodi attivi.
//
// Logica generale:
//   - Analizza i nodi attivi e li ordina per assurance crescente.
//   - Calcola score, assurance, e stato di ciascun nodo.
//   - Applica DP_findEligibleSolution per individuare le soluzioni “minime” eleggibili (più piccole e con minima assurance).
//   - Se nessuna soluzione è trovata, valuta anche i nodi Idle con un sovrapprezzo energetico.
//   - Genera soluzioni vicine (_DP_search_neigh_solutions) e seleziona quella col miglior punteggio globale.
//
// Parametri:
//   - cluster: snapshot del sistema.
//   - pod: Pod da posizionare.
//   - hyperparams.MaxReplicas: limite superiore di repliche (-1 = illimitato).
//   - scoringFunction: funzione di scoring per singolo nodo.
//   - explainScoringFunc: versione esplicativa della funzione di scoring.
//
// Ritorna:
//   - Una Solution accettata o rifiutata.
//
// Note:
//   - L’aggiunta del costo energetico ai nodi Idle penalizza la loro selezione, favorendo nodi già Attivi e riducendo il delta energetico.
//   - Gli array assurances, scores e references devono rimanere coerenti tra loro per indice.
//   - L’ordinamento per assurance crescente è fondamentale per la correttezza del DP.
func AddingNewPodDPStateAware(cluster *Cluster,
	pod *Pod,
	scoringFunction func(*WorkerNode, *Pod) float32,
	explainScoringFunc func(*WorkerNode, *Pod) map[string]float32,
	outputsAmount int,
	hyperparams AlgorithmHyperparams,
) []Solution {

	var solution Solution = InitSolution(pod)
	var edge_solutions map[int][]int

	// It is my responsability to create the vectors for assurances, scores and references
	var scores []float32
	var assurances []float32 // Probabilities
	var references []*WorkerNode
	var clusterstates []ClusterNodeState
	var n int
	var theta = pod.Criticality

	if len(cluster.Active) > 0 {
		/* Prepara vettori paralleli (score, assurance, reference, state) */
		if true { //Writing the arrays of scores and assurance. If true just to fold it
			for _, node := range cluster.Active {
				if node.EligibleFor(pod) {
					// Filtering for eligibles
					scores = append(scores, scoringFunction(node, pod))
					assurances = append(assurances, node.Assurance)
					references = append(references, node)
					clusterstates = append(clusterstates, Active)

				} else {
					// explain, reason := node.ExplainEligibility(pod)
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
		sortByPrimary_Assurance(assurances, scores, references, clusterstates, func(a, b float32) bool {
			return a < b
		}, false)
		/* Fase 1: ricerca soluzione minima eleggibile */
		_, edge_solutions = DP_findEligibleSolution(n, assurances, theta, hyperparams.MaxReplicas, hyperparams.DP_SolutionOversizeSearch)
	}

	// If no solution (that has theta or more) using only Active nodes, Try again on all
	if len(edge_solutions) == 0 {
		if true { //Writing the arrays of scores and assurance. If true just to fold it
			var overprice_func = func(true_score float32, node WorkerNode) float32 {
				return true_score + (hyperparams.DP_EnergyCostWakeupMultiplier * (float32(node.EnergyCostIdle) / float32(node.EnergyCostCoef))) + hyperparams.DP_EnergyCostWakeupConst
			}

			for _, node := range cluster.Idle {
				if node.EligibleFor(pod) {
					// Filtering for eligibles
					scores = append(scores, overprice_func(scoringFunction(node, pod), *node))

					assurances = append(assurances, node.Assurance)
					references = append(references, node)
					clusterstates = append(clusterstates, Idle)

				} else {
					// explain, reason := node.ExplainEligibility(pod)
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
		sortByPrimary_Assurance(assurances, scores, references, clusterstates, func(a, b float32) bool {
			return a < b
		}, false)
		_, edge_solutions = DP_findEligibleSolution(n, assurances, theta, hyperparams.MaxReplicas, hyperparams.DP_SolutionOversizeSearch)

		/*If still no solution, reject*/
		if len(edge_solutions) == 0 {
			solution.Reject("Unable to find eligible solutions")
			return []Solution{solution}
		}
	}

	//Debug
	// fmt.Printf("[DEBUG] probabilities %v\n", assurances)
	// fmt.Printf("[DEBUG] referecnes\n")
	// for i, r := range references {
	// 	fmt.Printf("\t%d\t%s\t%.2f\n", i, r.ID, r.Assurance)
	// }

	// Almeno una soluzione trovata
	/* Fase 2: espansione e ricerca di soluzioni adiacenti */
	all_eligibles := _DP_search_neigh_solutions(n, assurances, theta, edge_solutions, hyperparams.DP_NeighborSpan, hyperparams.DP_MaxNeighbors)

	/* Fase 3: Scegli la migliore e crea la struttura Solution */
	// fmt.Printf("\n\n[DEBUG] probabilities: %v\n\n", assurances)
	topSolutions, _ := _DP_pick_top_solutions(all_eligibles, scores, outputsAmount, hyperparams.DP_ScoreAggregationMode)
	var retSolutions []Solution = []Solution{}
	
	for _, tupSolution := range topSolutions{
		solution = _DP_tuple_to_solution(pod, tupSolution, references, assurances, scores, explainScoringFunc)
		solution.WrapUpSolution(cluster, hyperparams)
		retSolutions = append(retSolutions, solution)
	}

	return retSolutions
}

// AddingNewPodDPStateAgnostic implementa l’approccio DP analizzando da subito tutti i nodi (Active + Idle).
//
// Logica generale:
//   - Considera sia i nodi attivi che gli idle, applicando un sovrapprezzo ai secondi.
//   - Esegue la ricerca DP per identificare insiemi di nodi che rispettano la soglia di probabilità.
//   - Genera tutte le soluzioni eleggibili e sceglie quella col punteggio minimo.
//
// Parametri identici a DPStateAware.
//
// Note:
//   - Non distingue tra stati, quindi esplora un dominio più grande (maggiore costo computazionale).
//   - Cerca sempre l'ottimo globale tra tutti i nodi. (maggiore qualità di solito).
func AddingNewPodDPStateAgnostic(cluster *Cluster,
	pod *Pod,
	scoringFunction func(*WorkerNode, *Pod) float32,
	explainScoringFunc func(*WorkerNode, *Pod) map[string]float32,
	outputsAmount int,
	hyperparams AlgorithmHyperparams,
) []Solution {

	var solution Solution = InitSolution(pod)
	var edge_solutions map[int][]int

	// It is my responsability to create the vectors for assurances, scores and references
	var scores []float32
	var assurances []float32 // Probabilities
	var references []*WorkerNode
	var clusterstates []ClusterNodeState
	var n int
	var theta = pod.Criticality

	if len(cluster.Active) > 0 {
		if true { //Writing the arrays of scores and assurance. If true just to fold it
			for _, node := range cluster.Active {
				if node.EligibleFor(pod) {
					// Filtering for eligibles
					scores = append(scores, scoringFunction(node, pod))
					assurances = append(assurances, node.Assurance)
					references = append(references, node)
					clusterstates = append(clusterstates, Active)

				} else {
					// explain, reason := node.ExplainEligibility(pod)
				}
			}
		}
	}

	if true { //Writing the arrays of scores and assurance. If true just to fold it
		var overprice_func = func(true_score float32, node WorkerNode) float32 {
			return true_score + (hyperparams.DP_EnergyCostWakeupMultiplier * (float32(node.EnergyCostIdle) / float32(node.EnergyCostCoef))) + hyperparams.DP_EnergyCostWakeupConst
		}

		for _, node := range cluster.Idle {
			if node.EligibleFor(pod) {
				// Filtering for eligibles
				scores = append(scores, overprice_func(scoringFunction(node, pod), *node))

				assurances = append(assurances, node.Assurance)
				references = append(references, node)
				clusterstates = append(clusterstates, Idle)

			} else {
				// explain, reason := node.ExplainEligibility(pod)
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
	sortByPrimary_Assurance(assurances, scores, references, clusterstates, func(a, b float32) bool {
		return a < b
	}, false)
	_, edge_solutions = DP_findEligibleSolution(n, assurances, theta, hyperparams.MaxReplicas, hyperparams.DP_SolutionOversizeSearch)

	/*If still no solution, reject*/
	if len(edge_solutions) == 0 {
		solution.Reject("Unable to find eligible solutions")
			return []Solution{solution}
	}

	//One or more solution found
	// fmt.Printf("\n\n[DEBUG] probabilities: %v\n\n", assurances)
	all_eligibles := _DP_search_neigh_solutions(n, assurances, theta, edge_solutions, hyperparams.DP_NeighborSpan, hyperparams.DP_MaxNeighbors)
	
	/* Fase 3: Scegli la migliore e crea la struttura Solution */
	// fmt.Printf("\n\n[DEBUG] probabilities: %v\n\n", assurances)
	topSolutions, _ := _DP_pick_top_solutions(all_eligibles, scores, outputsAmount, hyperparams.DP_ScoreAggregationMode)
	var retSolutions []Solution = []Solution{}
	
	for _, tupSolution := range topSolutions{
		solution = _DP_tuple_to_solution(pod, tupSolution, references, assurances, scores, explainScoringFunc)
		solution.WrapUpSolution(cluster, hyperparams)
		retSolutions = append(retSolutions, solution)
	}

	return retSolutions
}

// AddingNewPodK8s implementa un algoritmo baseline ispirato alla logica “LeastAllocated” di Kubernetes.
//
// Logica generale:
//   - Seleziona iterativamente il nodo con punteggio migliore tra tutti i nodi disponibili.
//   - Aggiunge nodi finché non raggiunge la criticality richiesta o il limite di repliche.
//
// Differenze rispetto a Greedy:
//   - Non distingue Active/Idle: tutti i nodi concorrono fin da subito.
//   - È concettualmente più semplice ma meno efficiente dal punto di vista energetico.
//
// Ritorna:
//   - Una Solution accettata o rifiutata.
func AddingNewPodK8s(cluster *Cluster,
	pod *Pod,
	scoringFunction func(*WorkerNode, *Pod) float32,
	explainScoringFunc func(*WorkerNode, *Pod) map[string]float32,
	outputsAmount int,
	hyperparams AlgorithmHyperparams,
) []Solution {
	var solution Solution = InitSolution(pod)
	var exclude_ids cmn.Set = make(cmn.Set)
	var computed_scores = map[int]float32{}

	var id int = -1
	var score float32

	var probabilities = []float32{}
	var prob_atleast_half float32 = -1. //was 0.
	var theta = float32(pod.Criticality)

	// Find best among ALL nodes
	var allNodes_map = cluster.All_map()

	for prob_atleast_half < theta {
		if hyperparams.MaxReplicas > 0 && solution.Replicas == hyperparams.MaxReplicas {
			solution.Reject(fmt.Sprintf("Exceeded number of replicas (%d), while criticality constraint yet to meet (%.3f / %.3f)", hyperparams.MaxReplicas, prob_atleast_half, theta))
			break
		}
		/* Search greedily the best node */
		id, score = find_best_wn(allNodes_map, pod,
			true, exclude_ids, computed_scores,
			scoringFunction, ScoreEvaluation_condition)

		if id < 0 {
			solution.Reject(fmt.Sprintf("No more nodes to evaluate, while criticality constraint yet to meet (%.3f / %.3f)", prob_atleast_half, theta))
			break
		}
		//Found node
		node, _ := cluster.GetNodeByID(id)

		exclude_ids.Add(id)

		probabilities = append(probabilities, node.Assurance)
		prob_atleast_half = cmn.Compute_probability_atLeastHalf(probabilities)

		explainScore := explainScoringFunc(node, pod)
		solution.AddToSolution(node, prob_atleast_half, score, explainScore)
	}

	//Wrap it up
	if solution.Accepted {
		solution.WrapUpSolution(cluster, hyperparams)
	}

	return []Solution{solution}
}
