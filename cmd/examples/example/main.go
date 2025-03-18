package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Wooderan/sec-downloader-go/pkg/sec"
)

func main() {
	// Create a new downloader
	// Replace with your company name and email address to comply with SEC's fair access policy
	downloader, err := sec.NewDownloader("YourCompanyName", "your.email@example.com", "")
	if err != nil {
		log.Fatalf("Failed to create downloader: %v", err)
	}

	// Print supported forms
	fmt.Println("Supported forms:")
	forms := downloader.GetSupportedForms()
	for i, form := range forms {
		if i > 0 && i%5 == 0 {
			fmt.Println()
		}
		fmt.Printf("%-10s", form)
	}
	fmt.Println("\n")

	// Example 1: Download the latest 10-K filing for Apple
	fmt.Println("Example 1: Download the latest 10-K filing for Apple")
	fmt.Println("This will download the index.html file which contains links to all documents in the filing.")
	count, err := downloader.Get("10-K", "AAPL", 1, nil, nil, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n\n", count)

	// Example 2: Download 8-K filings for Microsoft from 2022-01-01 to 2022-12-31
	fmt.Println("Example 2: Download 8-K filings for Microsoft from 2022-01-01 to 2022-12-31")
	startDate := "2022-01-01"
	endDate := "2022-12-31"
	count, err = downloader.Get("8-K", "MSFT", 0, startDate, endDate, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n\n", count)

	// Example 3: Download 10-Q filings for Tesla, including amendments
	fmt.Println("Example 3: Download 10-Q filings for Tesla, including amendments")
	count, err = downloader.Get("10-Q", "TSLA", 5, nil, nil, true, false, nil)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n\n", count)

	// Example 4: Download 4 filings for a specific CIK (Elon Musk)
	fmt.Println("Example 4: Download Form 4 filings for a specific CIK (Elon Musk)")
	// Elon Musk's CIK
	elonMuskCIK := "0001494730"
	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	count, err = downloader.Get("4", elonMuskCIK, 10, oneMonthAgo, nil, false, true, nil)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n\n", count)

	// Print the location where files were saved
	cwd, _ := os.Getwd()
	fmt.Printf("Files were saved to: %s/%s\n", cwd, sec.RootSaveFolderName)

	// Example 5: Using the new options pattern
	fmt.Println("\nExample 5: Using the new options pattern to download 10-K filings for Nvidia")
	count, err = downloader.GetWithOptions(
		"10-K",
		"NVDA",
		sec.WithLimit(2),
		sec.WithDateRange("2021-01-01", nil),
		sec.WithIncludeAmends(true),
	)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n", count)
}
