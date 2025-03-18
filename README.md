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

	// Download the latest 10-K filing for Apple
	count, err := downloader.Get("10-K", "AAPL", 1, nil, nil, false, false, nil)
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

### `Get(form, tickerOrCIK string, limit int, after, before interface{}, includeAmends, downloadDetails bool, accessionNumbersToSkip map[string]bool) (int, error)`

Downloads filings and returns the number of filings downloaded.

- `form`: Form type to download (e.g., "8-K", "10-K")
- `tickerOrCIK`: Ticker or CIK for which to download filings
- `limit`: Max number of filings to download (0 for all available filings)
- `after`: Date after which to download filings (string in "YYYY-MM-DD" format or time.Time)
- `before`: Date before which to download filings (string in "YYYY-MM-DD" format or time.Time)
- `includeAmends`: Whether to include filing amendments (e.g., "8-K/A")
- `downloadDetails`: Whether to download filing details
- `accessionNumbersToSkip`: Map of accession numbers to skip

### `GetSupportedForms() []string`

Returns a list of supported form types.

## Examples

See the [examples](cmd/examples) for more usage examples.

## SEC Fair Access Policy

The SEC has a [fair access policy](https://www.sec.gov/os/webmaster-faq#code-support) that requires all programmatic access to include a company name and email address in the User-Agent header. This package complies with this policy by requiring these parameters when creating a `Downloader` instance.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 