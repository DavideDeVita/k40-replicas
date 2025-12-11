package core

type PodResources struct {
	Requested   int `json:"requested"` // currently requested
	Limit       int `json:"limit"`     // upper bound (used_limit)
}

type Pod struct {
	ID          string        `json:"id"`
	RealTime    bool
	CPU     PodResources `json:"cpu"`
	Memory  PodResources `json:"memory"`
	Storage PodResources `json:"storage"`
	Criticality float32
}