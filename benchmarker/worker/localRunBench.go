package main

import (
	"fmt"
	"os"
	"worker/score"
)

func localTest() {
	result, err := score.ComputeFinalScoreFromK6(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing final score: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Final Score: %d\n", result)
}
