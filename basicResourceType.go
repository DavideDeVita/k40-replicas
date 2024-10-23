package main

import (
	"fmt"
)

const (
	CPU  = iota // 0
	Disk        // 1
	RAM         // 2
)

/* Worker Node */
type BasicResourceType struct {
	request int
	limit   int
}

func (b BasicResourceType) Copy() BasicResourceType {
	return BasicResourceType{
		request: b.request,
		limit:   b.limit,
	}
}

func (br BasicResourceType) String() string {
	return fmt.Sprintf("r: %d   l: %d", br.request, br.limit)
}
