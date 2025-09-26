// worker/cmd/scorefile/main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"worker/score"
)

func main() {
	// go run calcScoreTest.go -f ../../scores/raw-data-127.0.0.1-20250911110719.json
	// jsonファイルにはrawデータを指定

	path := flag.String("f", "", "path to k6 summary-export JSON")
	flag.Parse()
	if *path == "" {
		log.Fatal("usage: scorefile -f /path/to/summary.json")
	}
	s, err := score.ComputeFinalScoreFromK6(*path)
	if err != nil {
		log.Fatalf("failed to compute score: %v", err)
	}
	fmt.Println(s)
}
