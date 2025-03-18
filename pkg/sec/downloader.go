package sec

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// DownloadOption represents an option for the Get method
type DownloadOption func(*DownloadMetadata)

// WithLimit sets the limit for the number of filings to download
func WithLimit(limit int) DownloadOption {
	return func(metadata *DownloadMetadata) {
		if limit > 0 {
			metadata.Limit = limit
		} else {
			metadata.Limit = math.MaxInt32
		}
	}
}

// WithDateRange sets the date range for filings to download
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

// WithIncludeAmends sets whether to include filing amendments
func WithIncludeAmends(includeAmends bool) DownloadOption {
	return func(metadata *DownloadMetadata) {
		metadata.IncludeAmends = includeAmends
	}
}

// WithDownloadDetails sets whether to download filing details
func WithDownloadDetails(downloadDetails bool) DownloadOption {
	return func(metadata *DownloadMetadata) {
		metadata.DownloadDetails = downloadDetails
	}
}

// WithAccessionNumbersToSkip sets accession numbers to skip
func WithAccessionNumbersToSkip(accessionNumbersToSkip map[string]bool) DownloadOption {
	return func(metadata *DownloadMetadata) {
		metadata.AccessionNumbersToSkip = accessionNumbersToSkip
	}
}

// Downloader is the main struct for downloading SEC filings
type Downloader struct {
	client         *SECClient
	downloadFolder string
	tickerToCIKMap map[string]string
}

// NewDownloader creates a new Downloader
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

// GetWithOptions downloads filings for a given form and ticker or CIK with options
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

// Get downloads filings for a given form and ticker or CIK
// This method is maintained for backward compatibility
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

// GetSupportedForms returns a list of supported forms
func (d *Downloader) GetSupportedForms() []string {
	forms := make([]string, 0, len(SupportedForms))
	for form := range SupportedForms {
		forms = append(forms, form)
	}
	return forms
}
