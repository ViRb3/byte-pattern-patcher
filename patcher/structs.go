package patcher

type patchDef struct {
	Label    string `json:"label,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Original string `json:"original"`
	Patched  string `json:"replaced"`
}

type Patch struct {
	Label             string
	Original          []byte
	OriginalWildcards []bool
	Patched           []byte
	PatchedWildcards  []bool
}

type parsedString struct {
	Pattern    []byte
	Wildcards  []bool
	Qualifiers []quantifier
}

type quantifier struct {
	Index int
	Min   int
	Max   int
}

type quantifierEx struct {
	Index  int
	Length int
}
