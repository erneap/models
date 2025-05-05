package systemdata

import "strings"

type Classification struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	SortID uint   `json:"sortID"`
}

type ByClassification []Classification

func (c ByClassification) Len() int { return len(c) }
func (c ByClassification) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByClassification) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type Communication struct {
	ID            string   `json:"id"`
	Explanation   string   `json:"explanation"`
	Exploitations []string `json:"exploitations,omitempty"`
	SortID        uint     `json:"sortID"`
}

func (c *Communication) HasExploitation(exp string) bool {
	for _, exploit := range c.Exploitations {
		if strings.EqualFold(exp, exploit) {
			return true
		}
	}
	return false
}

type ByCommunication []Communication

func (c ByCommunication) Len() int { return len(c) }
func (c ByCommunication) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByCommunication) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type DCGS struct {
	ID            string   `json:"id"`
	Exploitations []string `json:"exploitations,omitempty"`
	SortID        uint     `json:"sortID"`
}

type ByDCGS []DCGS

func (c ByDCGS) Len() int { return len(c) }
func (c ByDCGS) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByDCGS) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type Exploitation struct {
	ID          string `json:"id"`
	Explanation string `json:"exploitation"`
	SortID      uint   `json:"sortID"`
}

type ByExploitation []Exploitation

func (c ByExploitation) Len() int { return len(c) }
func (c ByExploitation) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByExploitation) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type SystemInfo struct {
	Classifications []Classification `json:"classifications,omitempty"`
	Communications  []Communication  `json:"communications,omitempty"`
	DCGSs           []DCGS           `json:"dCGSs,omitempty"`
	Exploitations   []Exploitation   `json:"exploitations,omitempty"`
	GroundSystems   []GroundSystem   `json:"groundSystems,omitempty"`
	Platforms       []Platform       `json:"platforms,omitempty"`
}

type SystemInfoResponse struct {
	SystemInfo SystemInfo `json:"systemInfo,omitempty"`
	Exception  string     `json:"exception,omitempty"`
}
