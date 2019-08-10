package main

func replacePattern(pattern Patch, array []byte, pos int) {
	for i := 0; i < len(pattern.Patched); i++ {
		if pos+i >= len(array) {
			return
		}
		if pattern.PatchedWildcards[i] {
			continue
		}
		array[pos+i] = pattern.Patched[i]
	}
}

func processPattern(pattern Patch, array []byte, length int) int {
	replaced := 0
	patternI := 0
	patternLen := len(pattern.Original)

	for i := 0; i < length; i++ {

		if pattern.OriginalWildcards[patternI] || pattern.Original[patternI] == array[i] {
			patternI++
		} else {
			// test: TestSequentialRepeating
			i -= patternI
			patternI = 0
			continue
		}

		if patternI == patternLen {
			// test: TestOffset
			replacePattern(pattern, array, i-patternLen+1)
			replaced++
			patternI = 0
			continue
		}
	}

	return replaced
}
