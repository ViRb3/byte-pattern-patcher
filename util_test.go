package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOffset(t *testing.T) {
	input := []byte{0x00, 0xAA, 0x00}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xAA},
			OriginalWildcards: []bool{false},
			Patched:           []byte{0xFF},
			PatchedWildcards:  []bool{false}},
		input)

	assert.Equal(t, []byte{0x00, 0xFF, 0x00}, input)
	assert.Equal(t, 1, replaced)
}

func TestSequentialRepeating(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0xAA},
			OriginalWildcards: []bool{false, false},
			Patched:           []byte{0xFF, 0xFF},
			PatchedWildcards:  []bool{false, false}},
		input)

	assert.Equal(t, []byte{0xEB, 0xFF, 0xFF}, input)
	assert.Equal(t, 1, replaced)
}

func TestOriginalWildcard(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0x00},
			OriginalWildcards: []bool{false, true},
			Patched:           []byte{0xFF, 0xFF},
			PatchedWildcards:  []bool{false, false}},
		input)

	assert.Equal(t, []byte{0xFF, 0xFF, 0xAA}, input)
	assert.Equal(t, 1, replaced)
}

func TestPatchWildcard(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0x00},
			OriginalWildcards: []bool{false, true},
			Patched:           []byte{0x00, 0xFF},
			PatchedWildcards:  []bool{true, false}},
		input)

	assert.Equal(t, []byte{0xEB, 0xFF, 0xAA}, input)
	assert.Equal(t, 1, replaced)
}

func TestTooLargePatch(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0xEB},
			OriginalWildcards: []bool{false, false},
			Patched:           []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			PatchedWildcards:  []bool{false, false, false, false, false, false}},
		input)

	assert.Equal(t, []byte{0x00, 0x11, 0x22}, input)
	assert.Equal(t, 1, replaced)
}
