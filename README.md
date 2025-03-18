# SEC Edgar Downloader for Go

A Go package for downloading company filings from the SEC EDGAR database. This is a port of the Python [sec-edgar-downloader](https://github.com/jadchaar/sec-edgar-downloader) package.

## Features

- Download filings for publicly traded companies, including their subsidiaries
- Download filings for individuals with a CIK
- Specify date ranges for filings
- Limit the number of filings to download
- Include filing amendments
- Download filing details (e.g., form 4 XML, 8-K HTML)
- Skip specific filings by accession number

## How It Works

This package downloads SEC filings by:

1. Fetching the ticker to CIK mapping from the SEC
2. Validating and converting the ticker or CIK
3. Fetching the list of available filings for the CIK
4. Filtering the filings based on form type, date range, etc.
5. Downloading the index.html file for each filing, which contains links to all documents in the filing
6. If a primary document is specified, downloading that document as well

The downloaded files are saved in a directory structure like:
```
sec-edgar-filings/
└── AAPL/
    └── 10-K/
        └── 0000320193-24-000123/
            ├── index.html
            └── primary-document.html
```

## Installation

```bash
go get github.com/user/sec-downloader-go
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/user/sec-downloader-go/pkg/sec"
)

func main() {
	// Create a new downloader
	// Replace with your company name and email address to comply with SEC's fair access policy
	downloader, err := sec.NewDownloader("YourCompanyName", "your.email@example.com", "")
	if err != nil {
		log.Fatalf("Failed to create downloader: %v", err)
	}

	// Download the latest 10-K filing for Apple using functional options pattern
	count, err := downloader.GetWithOptions("10-K", "AAPL", sec.WithLimit(1))
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n", count)
}
```

### Using Options Pattern

The options pattern provides a more flexible and readable way to configure the downloader:

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/user/sec-downloader-go/pkg/sec"
)

func main() {
	downloader, err := sec.NewDownloader("YourCompanyName", "your.email@example.com", "downloads")
	if err != nil {
		log.Fatalf("Failed to create downloader: %v", err)
	}

	// Download 10-Q filings for Tesla with multiple options
	count, err := downloader.GetWithOptions(
		"10-Q", 
		"TSLA",
		sec.WithLimit(5),
		sec.WithDateRange("2022-01-01", "2023-12-31"),
		sec.WithIncludeAmends(true),
		sec.WithDownloadDetails(true),
	)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n", count)

	// You can also use time.Time objects for date ranges
	startDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	
	count, err = downloader.GetWithOptions(
		"8-K", 
		"MSFT",
		sec.WithLimit(10),
		sec.WithDateRange(startDate, endDate),
	)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n", count)
	
	// Skip specific filings by accession number
	accessionNumbersToSkip := map[string]bool{
		"0001193125-22-067124": true,
		"0001193125-22-123456": true,
	}
	
	count, err = downloader.GetWithOptions(
		"10-K", 
		"AAPL",
		sec.WithAccessionNumbersToSkip(accessionNumbersToSkip),
	)
	if err != nil {
		log.Fatalf("Failed to download filings: %v", err)
	}
	fmt.Printf("Downloaded %d filings\n", count)
}
```

## API

### `NewDownloader(companyName, emailAddress string, downloadFolder string) (*Downloader, error)`

Creates a new `Downloader` instance.

- `companyName`: Your company name (required by SEC)
- `emailAddress`: Your email address (required by SEC)
- `downloadFolder`: Path to download location (defaults to current working directory)

### `GetWithOptions(form, tickerOrCIK string, options ...DownloadOption) (int, error)`

Downloads filings using the functional options pattern and returns the number of filings downloaded.

- `form`: Form type to download (e.g., "8-K", "10-K")
- `tickerOrCIK`: Ticker or CIK for which to download filings
- `options`: Variadic list of options to configure the download

### Option Functions

- `WithLimit(limit int)`: Sets the maximum number of filings to download (0 for all available)
- `WithDateRange(after, before interface{})`: Sets the date range for filings (string "YYYY-MM-DD" or time.Time)
- `WithIncludeAmends(includeAmends bool)`: Sets whether to include filing amendments
- `WithDownloadDetails(downloadDetails bool)`: Sets whether to download filing details
- `WithAccessionNumbersToSkip(accessionNumbersToSkip map[string]bool)`: Sets accession numbers to skip

### `Get(form, tickerOrCIK string, limit int, after, before interface{}, includeAmends, downloadDetails bool, accessionNumbersToSkip map[string]bool) (int, error)`

Legacy method that downloads filings and returns the number of filings downloaded.

### `GetSupportedForms() []string`

Returns a list of supported form types.

## Testing

The package includes both unit tests and integration tests. Unit tests can be run without making actual API calls to the SEC EDGAR database, while integration tests make real API calls.

### Running Unit Tests

To run only the unit tests, which don't make any actual SEC API calls:

```bash
cd pkg/sec && go test -v
```

### Running Integration Tests

Integration tests are disabled by default to avoid hitting SEC rate limits during CI. To run the integration tests:

```bash
cd pkg/sec && go test -v -integration
```

The integration tests will make real API calls to the SEC EDGAR database, so they should be run sparingly. The integration tests use a small number of requests and clean up after themselves.

## Examples

See the [examples](cmd/examples) for more usage examples.

## SEC Fair Access Policy

The SEC has a [fair access policy](https://www.sec.gov/os/webmaster-faq#code-support) that requires all programmatic access to include a company name and email address in the User-Agent header. This package complies with this policy by requiring these parameters when creating a `Downloader` instance.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 