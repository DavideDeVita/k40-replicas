package main

type Test struct {
	name      string
	Names     []string
	Callables []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution
	Scoring   []func(*WorkerNode, *Pod) float32

	MultiAware *MultiAwareParams
}

var (
	TEST_LeastAllocated = Test{
		name:      "leastAllocated",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_leastAllocated"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{costAware_leastAllocated_score, costAware_leastAllocated_score, k8s_leastAllocated_score},

		MultiAware: nil,
	}
	TEST_MostAllocated = Test{
		name:      "mostAllocated",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_mostAllocated"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{costAware_mostAllocated_score, costAware_mostAllocated_score, k8s_mostAllocated_score},

		MultiAware: nil,
	}
	TEST_RequestedToCapacityRatio = Test{
		name:      "requestedToCapacityRatio",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_requestedToCapacityRatio"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{costAware_requestedToCapacityRatio_score, costAware_requestedToCapacityRatio_score, k8s_requestedToCapacityRatio_score},

		MultiAware: nil,
	}
)

/*Multi Aware Params*/
type MultiAwareParams struct {
	k8s_func    func(*WorkerNode, *Pod) float32
	la_k8s_func func(*WorkerNode, *Pod) float32
	active      []bool
	weights     []float32
}

var (
	TEST_LeastAllocated_4Params = Test{
		name:      "leastAllocated_MAP",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_leastAllocated"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{parametric_scores_aware_score, parametric_scores_aware_score, k8s_leastAllocated_score},

		MultiAware: &MultiAwareParams{
			k8s_func:    k8s_leastAllocated_score,
			la_k8s_func: la_leastAllocated_score,
			active:      []bool{true, true, true, true},
			weights:     []float32{2, 1, 1, 1},
		},
	}
	TEST_MostAllocated_4Params = Test{
		name:      "mostAllocated_MAP",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_mostAllocated"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{parametric_scores_aware_score, parametric_scores_aware_score, k8s_mostAllocated_score},

		MultiAware: &MultiAwareParams{
			k8s_func:    k8s_mostAllocated_score,
			la_k8s_func: la_mostAllocated_score,
			active:      []bool{true, true, true, true},
			weights:     []float32{2, 1, 1, 1},
		},
	}
	TEST_RequestedToCapacityRatio_3Params = Test{
		name:      "requestedToCapacityRatio_MAP",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_requestedToCapacityRatio"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{parametric_scores_aware_score, parametric_scores_aware_score, k8s_requestedToCapacityRatio_score},

		MultiAware: &MultiAwareParams{
			k8s_func:    k8s_requestedToCapacityRatio_score,
			la_k8s_func: nil,
			active:      []bool{true, true, true, false},
			weights:     []float32{2, 1, 1, 1},
		},
	}
)

var (
	TEST_LA_LeastAllocated = Test{
		name:      "leastAllocated_LA",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_leastAllocated"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{parametric_scores_aware_score, parametric_scores_aware_score, k8s_leastAllocated_score},

		MultiAware: &MultiAwareParams{
			k8s_func:    k8s_leastAllocated_score, //Unused
			la_k8s_func: la_leastAllocated_score,
			active:      []bool{false, true, false, true},
			weights:     []float32{0, 1, 0, 2},
		},
	}
	TEST_LA_MostAllocated = Test{
		name:      "mostAllocated_LA",
		Names:     []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_mostAllocated"},
		Callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
		Scoring:   []func(*WorkerNode, *Pod) float32{parametric_scores_aware_score, parametric_scores_aware_score, k8s_mostAllocated_score},

		MultiAware: &MultiAwareParams{
			k8s_func:    k8s_mostAllocated_score, //Unused
			la_k8s_func: la_mostAllocated_score,
			active:      []bool{false, true, false, true},
			weights:     []float32{0, 1, 0, 2},
		},
	}
)
