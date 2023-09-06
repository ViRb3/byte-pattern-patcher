package main

import (
	"flag"
	"fmt"
	"github.com/ViRb3/byte-pattern-patcher/patcher"
	"io"
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

	patches, err := patcher.ReadPatches(*patchFileName)
	if err != nil {
		fmt.Println("read patches error: " + err.Error())
		os.Exit(1)
	}

	targetFile, err := os.OpenFile(*targetFileName, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("read file error: " + err.Error())
		os.Exit(1)
	}
	defer targetFile.Close()

	if *backup {
		if err := backupFile(targetFile); err != nil {
			fmt.Println("backup error: " + err.Error())
		}
	}

	results, err := patcher.Process(targetFile, patches)
	if err != nil {
		fmt.Println("patch error: " + err.Error())
		os.Exit(1)
	}

	replacedTotal := 0
	for pattern, count := range results {
		fmt.Printf("Replaced %d occurrences of %s\n", count, pattern)
		replacedTotal += count
	}

	fmt.Println("Done!")
	fmt.Printf("Replaced %d occurrences in total\n", replacedTotal)

	if replacedTotal < 1 {
		os.Exit(2)
	}
}
