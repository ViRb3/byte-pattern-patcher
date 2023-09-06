package patcher

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
)

func ReadPatches(patchFile string) ([]Patch, error) {
	bytes, err := os.ReadFile(patchFile)
	if err != nil {
		return nil, errors.New("read error: " + err.Error())
	}

	var result []PatchDef
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, errors.New("unmarshal error: " + err.Error())
	}

	parsedResult, err := ParsePatchDefs(result)
	if err != nil {
		return nil, errors.New("parse error: " + err.Error())
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

func ParsePatchDefs(patches []PatchDef) ([]Patch, error) {
	var result []Patch

	for _, patch := range patches {
		if patch.Disabled {
			continue
		}
		original, err := parseString(patch.Original)
		if err != nil {
			return nil, err
		}
		replaced, err := parseString(patch.Replaced)
		if err != nil {
			return nil, err
		}

		if len(original.Qualifiers) != len(replaced.Qualifiers) {
			return nil, errors.New(fmt.Sprintf("original quantifier len %d != replaced quantifier len %d",
				len(original.Qualifiers), len(replaced.Qualifiers)))
		}
		for i, q1 := range original.Qualifiers {
			q2 := replaced.Qualifiers[i]
			if q1.Index != q2.Index {
				return nil, errors.New(fmt.Sprintf("quantifier %d has mismatching index: %d and %d", i, q1.Index, q2.Index))
			}
			if q1.Min != q2.Min {
				return nil, errors.New(fmt.Sprintf("quantifier %d has mismatching mins: %d and %d", i, q1.Min, q2.Min))
			}
			if q1.Max != q2.Max {
				return nil, errors.New(fmt.Sprintf("quantifier %d has mismatching maxes: %d and %d", i, q1.Max, q2.Max))
			}
			if q1.Min < 1 {
				return nil, errors.New(fmt.Sprintf("quantifier %d has min < 1: %d", i, q1.Min))
			}
			if q1.Max < q1.Min {
				return nil, errors.New(fmt.Sprintf("quantifier %d has max < min: %d < %d", i, q1.Max, q1.Min))
			}
		}

		var expandedQuantifiers [][]interface{}
		for _, q := range original.Qualifiers {
			var expandedQuantifier []interface{}
			for qLen := q.Min; qLen <= q.Max; qLen++ {
				expandedQuantifier = append(expandedQuantifier, quantifierEx{q.Index, qLen})
			}
			expandedQuantifiers = append(expandedQuantifiers, expandedQuantifier)
		}

		for quantifierSet := range iter(expandedQuantifiers...) {
			originalPattern := original.Pattern
			originalWildcards := original.Wildcards
			replacedPattern := replaced.Pattern
			replacedWildcards := replaced.Wildcards

			// sort quantifiers starting with last so that we don't affect other quantifiers when we start expanding
			sort.Slice(quantifierSet, func(i, j int) bool {
				return quantifierSet[i].(quantifierEx).Index > quantifierSet[j].(quantifierEx).Index
			})

			for _, qInterface := range quantifierSet {
				q := qInterface.(quantifierEx)
				expandQuantifiersPattern(&originalPattern, q.Index, q.Length)
				expandQuantifiersWildcard(&originalWildcards, q.Index, q.Length)
				expandQuantifiersPattern(&replacedPattern, q.Index, q.Length)
				expandQuantifiersWildcard(&replacedWildcards, q.Index, q.Length)
			}

			result = append(result, Patch{
				Label:             patch.Label,
				Original:          originalPattern,
				OriginalWildcards: originalWildcards,
				Replaced:          replacedPattern,
				ReplacedWildcards: replacedWildcards})
		}
	}

	return result, nil
}

func parseString(pattern string) (parsedString, error) {
	elements := separatePattern.Split(pattern, -1)
	result := parsedString{
		Pattern:   make([]byte, len(elements)),
		Wildcards: make([]bool, len(elements)),
	}

	for i, element := range elements {
		matches := quantifierPattern.FindStringSubmatch(element)
		if matches != nil {
			// element with quantifier
			min, err := strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				return parsedString{}, err
			}
			max, err := strconv.ParseInt(matches[3], 10, 64)
			if err != nil {
				return parsedString{}, err
			}
			result.Qualifiers = append(result.Qualifiers, quantifier{
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
				return parsedString{}, err
			}
			result.Pattern[i] = parsed[0]
		}
	}

	return result, nil
}
