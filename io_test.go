package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuantifiers(t *testing.T) {
	patches, err := parsePatches([]PatchDef{{
		Label:    "test",
		Original: "01 ??{1,2} 02{1,2}",
		Patched:  "?? ??{1,2} 03{1,2}",
	}})
	if err != nil {
		t.Error(err)
	}
	expected := PatchData{
		Patches: []Patch{
			{
				Label:             "test",
				Original:          []byte{0x1, 0x0, 0x2},
				OriginalWildcards: []bool{false, true, false},
				Patched:           []byte{0x0, 0x0, 0x3},
				PatchedWildcards:  []bool{true, true, false},
			},
			{
				Label:             "test",
				Original:          []byte{0x1, 0x0, 0x2, 0x2},
				OriginalWildcards: []bool{false, true, false, false},
				Patched:           []byte{0x0, 0x0, 0x3, 0x3},
				PatchedWildcards:  []bool{true, true, false, false},
			},
			{
				Label:             "test",
				Original:          []byte{0x1, 0x0, 0x0, 0x2},
				OriginalWildcards: []bool{false, true, true, false},
				Patched:           []byte{0x0, 0x0, 0x0, 0x3},
				PatchedWildcards:  []bool{true, true, true, false},
			},
			{
				Label:             "test",
				Original:          []byte{01, 00, 00, 02, 02},
				OriginalWildcards: []bool{false, true, true, false, false},
				Patched:           []byte{0x0, 0x0, 0x0, 0x3, 0x3},
				PatchedWildcards:  []bool{true, true, true, false, false},
			},
		},
		LongestLen: 5,
	}
	assert.Equal(t, expected, patches)
}
