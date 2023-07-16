package main

import (
	"flag"
	"fmt"
	"github.com/ViRb3/byte-pattern-patcher/patcher"
	"os"
)

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
		fmt.Println(err)
		os.Exit(1)
	}

	results, err := patcher.Process(*targetFileName, *backup, patches)
	if err != nil {
		fmt.Println(err)
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
