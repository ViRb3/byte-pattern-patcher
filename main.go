package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func backupFile(file *os.File) error {
	backupFileName := file.Name() + ".bak"
	if _, err := os.Stat(backupFileName); err == nil {
		// assume backup exists
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	backupFile, err := os.OpenFile(backupFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, file)
	if err != nil {
		log.Fatal(err)
	}
	file.Seek(0, os.SEEK_SET)
	return nil
}

func main() {
	patchFileName := flag.String("p", "", "Patch definition")
	targetFileName := flag.String("t", "", "Target file")
	flag.Parse()

	if *patchFileName == "" || *targetFileName == "" {
		flag.Usage()
		return
	}

	patchData, err := readPatches(*patchFileName)
	if err != nil {
		log.Fatal(err)
	}

	targetFile, err := os.OpenFile(*targetFileName, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer targetFile.Close()

	if err := backupFile(targetFile); err != nil {
		log.Fatal(err)
	}

	bufferSize := 4096
	data := make([]byte, bufferSize)
	for {
		n, err := targetFile.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		replacedAll := 0
		for _, pattern := range patchData.Patches {
			replaced := processPattern(pattern, data, n)
			if replaced > 0 {
				if pattern.Label == "" {
					pattern.Label = "unlabeled"
				}
				fmt.Printf("Replaced %d occurances of %s\n", replaced, pattern.Label)
				replacedAll += replaced
			}
		}

		if replacedAll > 0 {
			targetFile.Seek(-int64(n), os.SEEK_CUR)
			nn, err := targetFile.Write(data[:n])
			if n != nn {
				log.Fatalf("Buffer size mismatch: %d vs %d", n, nn)
			}
			if err != nil {
				log.Fatal(err)
			}
		}

		// last buffer, reached end
		if n < bufferSize {
			break
		}

		// make sure we don't miss a pattern split between two buffers
		_, err = targetFile.Seek(-int64(patchData.LongestLen)-1, os.SEEK_CUR)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Done!")
}
