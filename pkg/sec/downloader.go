package sec

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// DownloadOption represents an option for the Get method.
// It's a function that modifies a DownloadMetadata instance.
// This follows the functional options pattern for configuring the downloader.
type DownloadOption func(*DownloadMetadata)

// WithLimit sets the limit for the number of filings to download.
// If limit is less than or equal to 0, it will download all available filings.
// Example: WithLimit(10) to download at most 10 filings.
func WithLimit(limit int) DownloadOption {
	return func(metadata *DownloadMetadata) {
		if limit > 0 {
			metadata.Limit = limit
		} else {
			metadata.Limit = math.MaxInt32
		}
	}
}

// WithDateRange sets the date range for filings to download.
// After and before can be either string in "YYYY-MM-DD" format or time.Time objects.
// Example: WithDateRange("2022-01-01", "2023-12-31")
func WithDateRange(after, before interface{}) DownloadOption {
	return func(metadata *DownloadMetadata) {
		if after != nil {
			if parsedAfter, err := ValidateAndParseDate(after); err == nil {
				metadata.After = parsedAfter
			}
		}
		if before != nil {
			if parsedBefore, err := ValidateAndParseDate(before); err == nil {
				metadata.Before = parsedBefore
			}
		}
	}
}

// WithIncludeAmends sets whether to include filing amendments.
// Filing amendments have form types with "/A" suffix (e.g., "10-K/A").
// Example: WithIncludeAmends(true) to include amendment filings.
func WithIncludeAmends(includeAmends bool) DownloadOption {
	return func(metadata *DownloadMetadata) {
		metadata.IncludeAmends = includeAmends
	}
}

// WithDownloadDetails sets whether to download filing details.
// Details are additional documents related to the filing.
// Example: WithDownloadDetails(true) to download filing details.
func WithDownloadDetails(downloadDetails bool) DownloadOption {
	return func(metadata *DownloadMetadata) {
		metadata.DownloadDetails = downloadDetails
	}
}

// WithAccessionNumbersToSkip sets accession numbers to skip during download.
// This is useful when you want to avoid re-downloading specific filings.
// Example: WithAccessionNumbersToSkip(map[string]bool{"0001193125-22-067124": true})
func WithAccessionNumbersToSkip(accessionNumbersToSkip map[string]bool) DownloadOption {
	return func(metadata *DownloadMetadata) {
		metadata.AccessionNumbersToSkip = accessionNumbersToSkip
	}
}

// Downloader is the main struct for downloading SEC filings.
// It provides methods to fetch and save SEC filings for companies and individuals.
type Downloader struct {
	client         *SECClient
	downloadFolder string
	tickerToCIKMap map[string]string
}

// NewDownloader creates a new Downloader instance.
//
// Parameters:
//   - companyName: Your company name (required by SEC fair access policy)
//   - emailAddress: Your email address (required by SEC fair access policy)
//   - downloadFolder: Path to download location (defaults to current working directory)
//
// Returns:
//   - A new Downloader instance and nil error on success
//   - nil and error on failure
//
// Example: NewDownloader("YourCompany", "your@email.com", "downloads")
func NewDownloader(companyName, emailAddress string, downloadFolder string) (*Downloader, error) {
	// Create the SEC client
	client := NewSECClient(companyName, emailAddress)

	// Set the download folder
	var folder string
	if downloadFolder == "" {
		// Use current working directory if no download folder is specified
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		folder = cwd
	} else {
		// Use the specified download folder
		absPath, err := filepath.Abs(downloadFolder)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		folder = absPath
	}

	// Get the ticker to CIK mapping
	tickerToCIKMap, err := client.GetTickerMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker to CIK mapping: %w", err)
	}

	return &Downloader{
		client:         client,
		downloadFolder: folder,
		tickerToCIKMap: tickerToCIKMap,
	}, nil
}

// GetWithOptions downloads filings for a given form and ticker or CIK with options.
// It uses the functional options pattern to configure the download.
//
// Parameters:
//   - form: Form type to download (e.g., "8-K", "10-K")
//   - tickerOrCIK: Ticker symbol or CIK for which to download filings
//   - options: Variadic list of options to configure the download
//
// Returns:
//   - Number of filings downloaded and nil error on success
//   - 0 and error on failure
//
// Example: GetWithOptions("10-K", "AAPL", WithLimit(5), WithDateRange("2022-01-01", "2023-12-31"))
func (d *Downloader) GetWithOptions(
	form string,
	tickerOrCIK string,
	options ...DownloadOption,
) (int, error) {
	// Check if the form is supported
	if !SupportedForms[form] {
		return 0, fmt.Errorf("form %s is not supported", form)
	}

	// Validate and convert the ticker or CIK
	cik, err := ValidateAndConvertTickerOrCIK(tickerOrCIK, d.tickerToCIKMap)
	if err != nil {
		return 0, fmt.Errorf("invalid ticker or CIK: %w", err)
	}

	// Create the download metadata with default values
	metadata := &DownloadMetadata{
		DownloadFolder:  d.downloadFolder,
		Form:            form,
		CIK:             cik,
		Limit:           math.MaxInt32,
		After:           DefaultAfterDate,
		Before:          DefaultBeforeDate,
		IncludeAmends:   false,
		DownloadDetails: false,
	}

	// Apply options
	for _, option := range options {
		option(metadata)
	}

	// If the ticker or CIK is a ticker, set it in the metadata
	if !IsCIK(tickerOrCIK) {
		metadata.Ticker = strings.ToUpper(tickerOrCIK)
	}

	// Fetch and save the filings
	return FetchAndSaveFilings(metadata, d.client)
}

// Get downloads filings for a given form and ticker or CIK.
// This method is maintained for backward compatibility.
//
// Parameters:
//   - form: Form type to download (e.g., "8-K", "10-K")
//   - tickerOrCIK: Ticker symbol or CIK for which to download filings
//   - limit: Max number of filings to download (0 for all available filings)
//   - after: Date after which to download filings (string in "YYYY-MM-DD" format or time.Time)
//   - before: Date before which to download filings (string in "YYYY-MM-DD" format or time.Time)
//   - includeAmends: Whether to include filing amendments (e.g., "8-K/A")
//   - downloadDetails: Whether to download filing details
//   - accessionNumbersToSkip: Map of accession numbers to skip
//
// Returns:
//   - Number of filings downloaded and nil error on success
//   - 0 and error on failure
//
// Example: Get("10-K", "AAPL", 5, "2022-01-01", "2023-12-31", false, true, nil)
func (d *Downloader) Get(
	form string,
	tickerOrCIK string,
	limit int,
	after interface{},
	before interface{},
	includeAmends bool,
	downloadDetails bool,
	accessionNumbersToSkip map[string]bool,
) (int, error) {
	options := []DownloadOption{
		WithLimit(limit),
		WithDateRange(after, before),
		WithIncludeAmends(includeAmends),
		WithDownloadDetails(downloadDetails),
	}

	if accessionNumbersToSkip != nil {
		options = append(options, WithAccessionNumbersToSkip(accessionNumbersToSkip))
	}

	return d.GetWithOptions(form, tickerOrCIK, options...)
}

// GetSupportedForms returns a list of supported form types.
//
// Returns:
//   - A slice of strings containing all supported form types
//
// Example:
//
//	forms := downloader.GetSupportedForms()
//	// forms contains ["10-K", "10-Q", "8-K", etc.]
func (d *Downloader) GetSupportedForms() []string {
	forms := make([]string, 0, len(SupportedForms))
	for form := range SupportedForms {
		forms = append(forms, form)
	}
	return forms
}
