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
		return err
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, file)
	if err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return nil
}

func main() {
	patchFileName := flag.String("p", "", "Patch definition file")
	targetFileName := flag.String("t", "", "Target file")
	backup := flag.Bool("b", true, "Backup original file")
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

	if *backup {
		if err := backupFile(targetFile); err != nil {
			log.Fatal(err)
		}
	}

	bufferSize := 4096
	data := make([]byte, bufferSize)
	replacedFile := 0
	for {
		n, err := targetFile.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		replacedBuffer := 0
		for _, pattern := range patchData.Patches {
			replacedPattern := processPattern(pattern, data[:n])
			if replacedPattern > 0 {
				if pattern.Label == "" {
					pattern.Label = "unlabeled"
				}
				fmt.Printf("Replaced %d occurrences of %s\n", replacedPattern, pattern.Label)
				replacedBuffer += replacedPattern
				replacedFile += replacedPattern
			}
		}

		if replacedBuffer > 0 {
			if _, err := targetFile.Seek(-int64(n), io.SeekCurrent); err != nil {
				log.Fatal(err)
			}
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
		_, err = targetFile.Seek(-int64(patchData.LongestLen)-1, io.SeekCurrent)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Done!")
	fmt.Printf("Replaced %d occurrences in total\n", replacedFile)

	if replacedFile < 1 {
		os.Exit(2)
	}
}
