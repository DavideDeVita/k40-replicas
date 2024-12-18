package main

type Test struct {
	name           string
	Names          []string
	Algo_callables []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution
	Is_multiparam  []bool

	Placing_scorer  func(*WorkerNode, *Pod) float32
	Placing_w       float32
	Multi_obj_funcs []func(*WorkerNode, *Pod) float32
	Multi_obj_w     []float32
}

// Classics
var TEST_LeastAllocated = Test{
	name:           "requested",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_requested"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, false},

	Placing_scorer:  k8s_leastAllocated_score,
	Placing_w:       1,
	Multi_obj_funcs: nil,
	Multi_obj_w:     nil,
}

var TEST_MostAllocated = Test{
	name:           "mostAllocated",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_mostAllocated"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, false},

	Placing_scorer:  k8s_mostAllocated_score,
	Placing_w:       1,
	Multi_obj_funcs: nil,
	Multi_obj_w:     nil,
}

var TEST_RequestedToCapacityRatio = Test{
	name:           "requestedToCapacityRatio",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_requestedToCapacityRatio"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, false},

	Placing_scorer:  k8s_requestedToCapacityRatio_score,
	Placing_w:       1,
	Multi_obj_funcs: nil,
	Multi_obj_w:     nil,
}

// 4 params
var TEST_LeastAllocated_4Params = Test{
	name:           "custom_test",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_requested"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, false},

	Placing_scorer:  k8s_leastAllocated_score,
	Placing_w:       2,
	Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _computationPower_ratio, _log10_assurance},
	Multi_obj_w:     []float32{1, 1, 1},
}

var TEST_MostAllocated_4Params = Test{
	name:           "custom_test",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_mostAllocated"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, false},

	Placing_scorer:  k8s_mostAllocated_score,
	Placing_w:       2,
	Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _computationPower_ratio, _log10_assurance},
	Multi_obj_w:     []float32{1, 1, 1},
}

var TEST_RequestedToCapacityRatio_4Params = Test{
	name:           "custom_test",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic", "K8s_requestedToCapacityRatio"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, false},

	Placing_scorer:  k8s_requestedToCapacityRatio_score,
	Placing_w:       2,
	Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _computationPower_ratio, _log10_assurance},
	Multi_obj_w:     []float32{1, 1, 1},
}
