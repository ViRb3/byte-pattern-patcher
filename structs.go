package main

type PatchDef struct {
	Label    string `json:"label,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
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

type ParsedString struct {
	Pattern    []byte
	Wildcards  []bool
	Qualifiers []Quantifier
}

type Quantifier struct {
	Index int
	Min   int
	Max   int
}

type QuantifierEx struct {
	Index  int
	Length int
}
