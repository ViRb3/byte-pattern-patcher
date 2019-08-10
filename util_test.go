package main

import (
	"bytes"
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
		input,
		len(input))

	if !bytes.Equal(input, []byte{0x00, 0xFF, 0x00}) {
		t.Error("test failed!")
	}
	if replaced != 1 {
		t.Error("test failed (2)!")
	}
}

func TestSequentialRepeating(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0xAA},
			OriginalWildcards: []bool{false, false},
			Patched:           []byte{0xFF, 0xFF},
			PatchedWildcards:  []bool{false, false}},
		input,
		len(input))

	if !bytes.Equal(input, []byte{0xEB, 0xFF, 0xFF}) {
		t.Error("test failed!")
	}
	if replaced != 1 {
		t.Error("test failed (2)!")
	}
}

func TestOriginalWildcard(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0x00},
			OriginalWildcards: []bool{false, true},
			Patched:           []byte{0xFF, 0xFF},
			PatchedWildcards:  []bool{false, false}},
		input,
		len(input))

	if !bytes.Equal(input, []byte{0xFF, 0xFF, 0xAA}) {
		t.Error("test failed!")
	}
	if replaced != 1 {
		t.Error("test failed (2)!")
	}
}

func TestPatchWildcard(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0x00},
			OriginalWildcards: []bool{false, true},
			Patched:           []byte{0x00, 0xFF},
			PatchedWildcards:  []bool{true, false}},
		input,
		len(input))

	if !bytes.Equal(input, []byte{0xEB, 0xFF, 0xAA}) {
		t.Error("test failed!")
	}
	if replaced != 1 {
		t.Error("test failed (2)!")
	}
}

func TestTooLargePatch(t *testing.T) {
	input := []byte{0xEB, 0xEB, 0xAA}
	replaced := processPattern(
		Patch{
			Original:          []byte{0xEB, 0xEB},
			OriginalWildcards: []bool{false, false},
			Patched:           []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			PatchedWildcards:  []bool{false, false, false, false, false, false}},
		input,
		len(input))

	if !bytes.Equal(input, []byte{0x00, 0x11, 0x22}) {
		t.Error("test failed!")
	}
	if replaced != 1 {
		t.Error("test failed (2)!")
	}
}
