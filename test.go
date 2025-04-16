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
	Multi_obj_names []string
}

// Classics
var TEST_LeastAllocated = Test{
	name:           "leastAllocated",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic byState", "K4.0 Dynamic ALL", "K8s_mostAllocated"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__dynamic_allNodes, adding_new_pod__k8s},
	Is_multiparam:  []bool{false, false, false, false},

	Placing_scorer:  k8s_leastAllocated_score,
	Placing_w:       1,
	Multi_obj_funcs: nil,
	Multi_obj_w:     nil,
	Multi_obj_names: nil,
}

var TEST_MostAllocated = Test{
	name:           "mostAllocated",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic byState", "K4.0 Dynamic ALL", "K8s_mostAllocated"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__dynamic_allNodes, adding_new_pod__k8s},
	Is_multiparam:  []bool{false, false, false, false},

	Placing_scorer:  k8s_mostAllocated_score,
	Placing_w:       1,
	Multi_obj_funcs: nil,
	Multi_obj_w:     nil,
	Multi_obj_names: nil,
}

var TEST_RequestedToCapacityRatio = Test{
	name:           "requestedToCapacityRatio",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic byState", "K4.0 Dynamic ALL", "K8s_mostAllocated"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__dynamic_allNodes, adding_new_pod__k8s},
	Is_multiparam:  []bool{false, false, false, false},

	Placing_scorer:  k8s_requestedToCapacityRatio_score,
	Placing_w:       1,
	Multi_obj_funcs: nil,
	Multi_obj_w:     nil,
	Multi_obj_names: nil,
}

// 4 params

// 5 params
var TEST_LeastAllocated_5Params = Test{
	name:           "leastAllocated",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic byState", "K4.0 Dynamic ALL", "K8s_leastAllocated"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__dynamic_allNodes, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, true, false},

	Placing_scorer:  k8s_leastAllocated_score,
	Placing_w:       4,
	Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _sigma_assurance_wasteless, _computationPower_ratio, _rt_waste},
	Multi_obj_w:     []float32{2, 2, 1, 1},
	Multi_obj_names: []string{"energy cost", "log assurance", "comput power", "rt waste"},
}

var TEST_MostAllocated_5Params = Test{
	name:           "mostAllocated",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic byState", "K4.0 Dynamic ALL", "K8s_mostAllocated"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__dynamic_allNodes, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, true, false},

	Placing_scorer:  k8s_mostAllocated_score,
	Placing_w:       4,
	Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _sigma_assurance_wasteless, _computationPower_ratio, _rt_waste},
	Multi_obj_w:     []float32{2, 2, 1, 1},
	Multi_obj_names: []string{"energy cost", "log assurance", "comput power", "rt waste"},
}

var TEST_RequestedToCapacityRatio_5Params = Test{
	name:           "requestedToCapacityRatio",
	Names:          []string{"K4.0 Greedy", "K4.0 Dynamic byState", "K4.0 Dynamic ALL", "K8s_requestedToCapacityRatio"},
	Algo_callables: []func(*Cluster, *Pod, func(*WorkerNode, *Pod) float32) Solution{adding_new_pod__greedy, adding_new_pod__dynamic, adding_new_pod__dynamic_allNodes, adding_new_pod__k8s},
	Is_multiparam:  []bool{true, true, true, false},

	Placing_scorer:  k8s_requestedToCapacityRatio_score,
	Placing_w:       4,
	Multi_obj_funcs: []func(*WorkerNode, *Pod) float32{_energyCost_ratio, _sigma_assurance_wasteless, _computationPower_ratio, _rt_waste},
	Multi_obj_w:     []float32{2, 2, 1, 1},
	Multi_obj_names: []string{"energy cost", "log assurance", "comput power", "rt waste"},
}
