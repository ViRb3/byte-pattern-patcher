package main

type PatchDef struct {
	Label    string `json:"label,omitempty"`
	Original string `json:"original"`
	Patched  string `json:"replaced"`
}

type PatchData struct {
	Patches    []Patch
	LongestLen int
}

type Patch struct {
	Label             string
	Original          []byte
	OriginalWildcards []bool
	Patched           []byte
	PatchedWildcards  []bool
}
