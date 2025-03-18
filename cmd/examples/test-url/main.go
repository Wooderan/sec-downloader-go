package main

import (
	"fmt"
	"log"

	"github.com/Wooderan/sec-downloader-go/pkg/sec"
)

func main() {
	// Create a new SEC client
	client := sec.NewSECClient("TestCompany", "test@example.com")

	// Example CIK and accession number for Apple
	cik := "0000320193"
	accNum := "0000320193-24-000123"

	// Create a ToDownload object
	td, err := sec.GetToDownload(cik, accNum, "")
	if err != nil {
		log.Fatalf("Failed to create ToDownload object: %v", err)
	}

	// Print the URLs
	fmt.Println("Raw Filing URI (index page):")
	fmt.Println(td.RawFilingURI)
	fmt.Println()

	// Try to download the raw filing
	fmt.Println("Downloading the index page...")
	rawFiling, err := client.DownloadFiling(td.RawFilingURI)
	if err != nil {
		log.Fatalf("Failed to download raw filing: %v", err)
	}

	// Print the size of the downloaded file
	fmt.Printf("Successfully downloaded %d bytes\n", len(rawFiling))

	// Print a preview of the content
	preview := string(rawFiling)
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	fmt.Printf("\nPreview of the content:\n%s\n", preview)
}
