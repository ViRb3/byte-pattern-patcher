package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
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
var quantifierPattern = regexp.MustCompile(`^([A-z0-9?]{1,2}){(\d+),(\d+)}$`)

func expandQuantifiersPattern(pattern *[]byte, startI int, qLen int) {
	newPattern := make([]byte, 0, len(*pattern)-1+qLen)
	newPattern = append(newPattern, (*pattern)[:startI]...)
	for i := startI; i < startI+qLen; i++ {
		newPattern = append(newPattern, (*pattern)[startI])
	}
	newPattern = append(newPattern, (*pattern)[startI+1:]...)
	*pattern = newPattern
}

func expandQuantifiersWildcard(pattern *[]bool, startI int, qLen int) {
	newPattern := make([]bool, 0, len(*pattern)-1+qLen)
	newPattern = append(newPattern, (*pattern)[:startI]...)
	for i := startI; i < startI+qLen; i++ {
		newPattern = append(newPattern, (*pattern)[startI])
	}
	newPattern = append(newPattern, (*pattern)[startI+1:]...)
	*pattern = newPattern
}

func parsePatches(patches []PatchDef) (PatchData, error) {
	result := PatchData{}

	for _, patch := range patches {
		if patch.Disabled {
			continue
		}
		original, err := parseString(patch.Original)
		if err != nil {
			return PatchData{}, err
		}
		patched, err := parseString(patch.Patched)
		if err != nil {
			return PatchData{}, err
		}

		if len(original.Qualifiers) != len(patched.Qualifiers) {
			return PatchData{}, errors.New(fmt.Sprintf("original quantifier len %d != patched quantifier len %d",
				len(original.Qualifiers), len(patched.Qualifiers)))
		}
		for i, q1 := range original.Qualifiers {
			q2 := patched.Qualifiers[i]
			if q1.Index != q2.Index {
				return PatchData{}, errors.New(fmt.Sprintf("quantifier %d has mismatching index: %d and %d", i, q1.Index, q2.Index))
			}
			if q1.Min != q2.Min {
				return PatchData{}, errors.New(fmt.Sprintf("quantifier %d has mismatching mins: %d and %d", i, q1.Min, q2.Min))
			}
			if q1.Max != q2.Max {
				return PatchData{}, errors.New(fmt.Sprintf("quantifier %d has mismatching maxes: %d and %d", i, q1.Max, q2.Max))
			}
			if q1.Min < 1 {
				return PatchData{}, errors.New(fmt.Sprintf("quantifier %d has min < 1: %d", i, q1.Min))
			}
			if q1.Max < q1.Min {
				return PatchData{}, errors.New(fmt.Sprintf("quantifier %d has max < min: %d < %d", i, q1.Max, q1.Min))
			}
		}

		var expandedQuantifiers [][]interface{}
		for _, q := range original.Qualifiers {
			var expandedQuantifier []interface{}
			for qLen := q.Min; qLen <= q.Max; qLen++ {
				expandedQuantifier = append(expandedQuantifier, QuantifierEx{q.Index, qLen})
			}
			expandedQuantifiers = append(expandedQuantifiers, expandedQuantifier)
		}

		for quantifierSet := range Iter(expandedQuantifiers...) {
			originalPattern := original.Pattern
			originalWildcards := original.Wildcards
			patchedPattern := patched.Pattern
			patchedWildcards := patched.Wildcards

			// sort quantifiers starting with last so that we don't affect other quantifiers when we start expanding
			sort.Slice(quantifierSet, func(i, j int) bool {
				return quantifierSet[i].(QuantifierEx).Index > quantifierSet[j].(QuantifierEx).Index
			})

			for _, qInterface := range quantifierSet {
				q := qInterface.(QuantifierEx)
				expandQuantifiersPattern(&originalPattern, q.Index, q.Length)
				expandQuantifiersWildcard(&originalWildcards, q.Index, q.Length)
				expandQuantifiersPattern(&patchedPattern, q.Index, q.Length)
				expandQuantifiersWildcard(&patchedWildcards, q.Index, q.Length)
			}

			result.Patches = append(result.Patches, Patch{
				Label:             patch.Label,
				Original:          originalPattern,
				OriginalWildcards: originalWildcards,
				Patched:           patchedPattern,
				PatchedWildcards:  patchedWildcards})

			longestLen := len(original.Pattern)
			for _, q := range original.Qualifiers {
				longestLen += q.Max - 1
			}
			if longestLen > result.LongestLen {
				result.LongestLen = longestLen
			}
		}
	}

	return result, nil
}

func parseString(pattern string) (ParsedString, error) {
	elements := separatePattern.Split(pattern, -1)
	result := ParsedString{
		Pattern:   make([]byte, len(elements)),
		Wildcards: make([]bool, len(elements)),
	}

	for i, element := range elements {
		matches := quantifierPattern.FindStringSubmatch(element)
		if matches != nil {
			// element with quantifier
			min, err := strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				return ParsedString{}, err
			}
			max, err := strconv.ParseInt(matches[3], 10, 64)
			if err != nil {
				return ParsedString{}, err
			}
			result.Qualifiers = append(result.Qualifiers, Quantifier{
				Index: i,
				Min:   int(min),
				Max:   int(max),
			})
			// remove the quantifier part, leaving just a plain byte or wildcard
			element = matches[1]
		}
		if element == "?" || element == "??" {
			result.Pattern[i] = 0x0
			result.Wildcards[i] = true
		} else {
			parsed, err := hex.DecodeString(element)
			if err != nil {
				return ParsedString{}, err
			}
			result.Pattern[i] = parsed[0]
		}
	}

	return result, nil
}
