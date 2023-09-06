package patcher

type PatchDef struct {
	Label    string `json:"label,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Original string `json:"original"`
	Replaced string `json:"replaced"`
}

type Patch struct {
	Label             string
	Original          []byte
	OriginalWildcards []bool
	Replaced          []byte
	ReplacedWildcards []bool
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
