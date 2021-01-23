package main

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"regexp"
)

func readPatches(patchFile string) (PatchData, error) {
	bytes, err := ioutil.ReadFile(patchFile)
	if err != nil {
		return PatchData{}, err
	}

	var result []PatchDef
	if err := json.Unmarshal(bytes, &result); err != nil {
		return PatchData{}, err
	}

	parsedResult, err := parsePatches(result)
	if err != nil {
		return PatchData{}, err
	}
	return parsedResult, nil
}

var separatePattern = regexp.MustCompile("\\s+")

func parsePatches(patches []PatchDef) (PatchData, error) {
	result := PatchData{}

	for _, patch := range patches {
		original, origWildcards, err := stringToByteArray(patch.Original)
		if err != nil {
			return PatchData{}, err
		}
		patched, patchWildcards, err := stringToByteArray(patch.Patched)
		if err != nil {
			return PatchData{}, err
		}
		newPatch := Patch{
			Label:             patch.Label,
			Original:          original,
			OriginalWildcards: origWildcards,
			Patched:           patched,
			PatchedWildcards:  patchWildcards}
		result.Patches = append(result.Patches, newPatch)

		if len(original) > result.LongestLen {
			result.LongestLen = len(origWildcards)
		}
	}

	return result, nil
}

func stringToByteArray(pattern string) ([]byte, []bool, error) {
	bytes := separatePattern.Split(pattern, -1)
	result := make([]byte, len(bytes))
	wildcards := make([]bool, len(bytes))

	for i, b := range bytes {
		if b == "?" || b == "??" {
			result[i] = 0x0
			wildcards[i] = true
		} else {
			parsed, err := hex.DecodeString(b)
			if err != nil {
				return nil, nil, err
			}
			result[i] = parsed[0]
		}
	}

	return result, wildcards, nil
}
