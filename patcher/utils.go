package patcher

import (
	"errors"
	"fmt"
	"io"
)

func applyPatchAtPos(patch Patch, array []byte, pos int) {
	for i := 0; i < len(patch.Replaced); i++ {
		if pos+i >= len(array) {
			return
		}
		if patch.ReplacedWildcards[i] {
			continue
		}
		array[pos+i] = patch.Replaced[i]
	}
}

func applyPatch(patch Patch, array []byte) int {
	replaced := 0
	patchPos := 0
	patchLen := len(patch.Original)

	for i := 0; i < len(array); i++ {

		if patch.OriginalWildcards[patchPos] || patch.Original[patchPos] == array[i] {
			patchPos++
		} else {
			// test: TestSequentialRepeating
			i -= patchPos
			patchPos = 0
			continue
		}

		if patchPos == patchLen {
			// test: TestOffset
			applyPatchAtPos(patch, array, i-patchLen+1)
			replaced++
			patchPos = 0
			continue
		}
	}

	return replaced
}

func Process(targetFile io.ReadWriteSeeker, patches []Patch) (map[string]int, error) {
	longestLen := 0
	for _, patch := range patches {
		if len(patch.Original) > longestLen {
			longestLen = len(patch.Original)
		}
	}
	bufferSize := 4096
	data := make([]byte, bufferSize)
	result := map[string]int{}
	for {
		n, err := targetFile.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.New("read error: " + err.Error())
		}

		replacedTotal := 0
		for _, pattern := range patches {
			replacedCount := applyPatch(pattern, data[:n])
			if replacedCount > 0 {
				if pattern.Label == "" {
					pattern.Label = "unlabeled"
				}
				if oldCount, ok := result[pattern.Label]; ok {
					result[pattern.Label] = oldCount + replacedCount
				} else {
					result[pattern.Label] = replacedCount
				}
				replacedTotal += replacedCount
			}
		}

		// save changed buffer to file
		if replacedTotal > 0 {
			if _, err := targetFile.Seek(-int64(n), io.SeekCurrent); err != nil {
				return nil, errors.New("seek error: " + err.Error())
			}
			nn, err := targetFile.Write(data[:n])
			if err != nil {
				return nil, errors.New("write error: " + err.Error())
			} else if n != nn {
				return nil, errors.New(fmt.Sprintf("buffer size mismatch: %d vs %d", n, nn))
			}
		}

		// last buffer, reached end
		if n < bufferSize {
			break
		}

		// make sure we don't miss a pattern split between two buffers
		_, err = targetFile.Seek(-int64(longestLen)-1, io.SeekCurrent)
		if err != nil {
			return nil, errors.New("seek error: " + err.Error())
		}
	}
	return result, nil
}
