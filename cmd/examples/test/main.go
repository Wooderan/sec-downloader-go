package main

import (
	"fmt"
	"log"

	"github.com/Wooderan/sec-downloader-go/pkg/sec"
)

func main() {
	// Create a new SEC client
	client := sec.NewSECClient("TestCompany", "test@example.com")

	// Try to get the ticker to CIK mapping
	fmt.Println("Fetching ticker to CIK mapping...")
	tickerToCIKMap, err := client.GetTickerMetadata()
	if err != nil {
		log.Fatalf("Failed to get ticker to CIK mapping: %v", err)
	}

	// Print the number of mappings found
	fmt.Printf("Found %d ticker to CIK mappings\n", len(tickerToCIKMap))

	// Print a few example mappings
	examples := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "META"}
	fmt.Println("\nExample mappings:")
	for _, ticker := range examples {
		cik, ok := tickerToCIKMap[ticker]
		if ok {
			fmt.Printf("%s -> %s\n", ticker, cik)
		} else {
			fmt.Printf("%s -> Not found\n", ticker)
		}
	}
}
