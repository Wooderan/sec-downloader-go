package sec

import (
	"flag"
	"os"
	"testing"
)

// Define a flag to control whether integration tests are run
var runIntegrationTests = flag.Bool("integration", false, "Run integration tests that make real API calls to SEC")

// TestMain is used to setup the testing environment
func TestMain(m *testing.M) {
	// Parse command line flags
	flag.Parse()

	// Run tests
	os.Exit(m.Run())
}

// TestIntegrationDownloader tests the actual downloading functionality with real SEC API calls
// To run this test, use: go test -v -integration
func TestIntegrationDownloader(t *testing.T) {
	// Skip if integration tests are not enabled
	if !*runIntegrationTests {
		t.Skip("Skipping integration tests. Use -integration flag to run them.")
	}

	// Create a real downloader
	downloader, err := NewDownloader("SEC-Go-Test", "test@example.com", "")
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Test fetching a single 10-K for a well-known company (limit=1)
	count, err := downloader.GetWithOptions("10-K", "AAPL", WithLimit(1))
	if err != nil {
		t.Errorf("GetWithOptions() for AAPL 10-K error: %v", err)
	} else {
		t.Logf("Successfully downloaded %d 10-K filing(s) for AAPL", count)
	}

	// Clean up downloaded files
	// This assumes files are saved in the current directory under sec-edgar-filings
	if err := os.RemoveAll("sec-edgar-filings"); err != nil {
		t.Logf("Warning: Failed to clean up test files: %v", err)
	}
}

// TestIntegrationTickerToCIK tests the ticker to CIK mapping functionality
func TestIntegrationTickerToCIK(t *testing.T) {
	// Skip if integration tests are not enabled
	if !*runIntegrationTests {
		t.Skip("Skipping integration tests. Use -integration flag to run them.")
	}

	// Create a client
	client := NewSECClient("SEC-Go-Test", "test@example.com")

	// Get the ticker to CIK mapping
	tickerToCIKMap, err := client.GetTickerMetadata()
	if err != nil {
		t.Fatalf("Failed to get ticker to CIK mapping: %v", err)
	}

	// Check for well-known companies
	expectedCompanies := []struct {
		ticker string
		cik    string
	}{
		{"AAPL", "0000320193"},
		{"MSFT", "0000789019"},
		{"GOOG", "0001652044"},
	}

	for _, company := range expectedCompanies {
		cik, ok := tickerToCIKMap[company.ticker]
		if !ok {
			t.Errorf("Ticker %s not found in mapping", company.ticker)
			continue
		}

		if cik != company.cik {
			t.Errorf("CIK for %s = %s, want %s", company.ticker, cik, company.cik)
		} else {
			t.Logf("Successfully mapped ticker %s to CIK %s", company.ticker, cik)
		}
	}
}
