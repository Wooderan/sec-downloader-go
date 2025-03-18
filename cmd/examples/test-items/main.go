package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Wooderan/sec-downloader-go/pkg/sec"
)

func main() {
	// Create a new SEC client
	client := sec.NewSECClient("TestCompany", "test@example.com")

	// Try to get filings for Apple (CIK: 0000320193)
	fmt.Println("Fetching 8-K filings for Apple (CIK: 0000320193)...")

	// Format the submission URL
	submissionFile := fmt.Sprintf(sec.SubmissionFileFormat, "0000320193")
	submissionURL := fmt.Sprintf(sec.URLSubmissions, submissionFile)

	// Get the list of available filings
	submissionData, err := client.GetListOfAvailableFilings(submissionURL)
	if err != nil {
		log.Fatalf("Failed to get list of available filings: %v", err)
	}

	// Get the recent filings
	recent := submissionData.Filings.Recent

	// Filter for 8-K filings
	var eightKIndices []int
	for i, form := range recent.Form {
		if form == "8-K" {
			eightKIndices = append(eightKIndices, i)
		}
	}

	// Print the number of 8-K filings found
	fmt.Printf("Found %d 8-K filings\n", len(eightKIndices))

	// Print a few example filings with their items
	fmt.Println("\nExample 8-K filings with items:")
	limit := 5
	if len(eightKIndices) < limit {
		limit = len(eightKIndices)
	}

	for i := 0; i < limit; i++ {
		idx := eightKIndices[i]
		fmt.Printf("Filing %d:\n", i+1)
		fmt.Printf("  Accession Number: %s\n", recent.AccessionNumber[idx])
		fmt.Printf("  Form: %s\n", recent.Form[idx])
		fmt.Printf("  Filing Date: %s\n", recent.FilingDate[idx])

		// Print the items (if any)
		if idx < len(recent.Items) && recent.Items[idx] != "" {
			fmt.Printf("  Items: %s\n", recent.Items[idx])

			// Split the items by comma and print each item
			items := strings.Split(recent.Items[idx], ",")
			fmt.Printf("  Parsed Items (%d):\n", len(items))
			for j, item := range items {
				fmt.Printf("    Item %d: %s\n", j+1, strings.TrimSpace(item))
			}
		} else {
			fmt.Printf("  Items: None\n")
		}

		fmt.Println()
	}
}
