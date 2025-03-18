package sec

import (
	"time"
)

// DownloadMetadata represents the metadata for a filing download operation.
// It contains all the parameters needed to configure a download request,
// including target form, CIK, date ranges, and various filtering options.
type DownloadMetadata struct {
	// DownloadFolder is the directory where filings will be saved
	DownloadFolder string
	// Form is the SEC form type to download (e.g., "10-K", "8-K")
	Form string
	// CIK is the Central Index Key that identifies the company
	CIK string
	// Limit is the maximum number of filings to download (0 for all available)
	Limit int
	// After is the date after which to consider filings
	After time.Time
	// Before is the date before which to consider filings
	Before time.Time
	// IncludeAmends determines whether to include filing amendments (e.g., "10-K/A")
	IncludeAmends bool
	// DownloadDetails determines whether to download filing details documents
	DownloadDetails bool
	// Ticker is the stock ticker symbol if available (optional)
	Ticker string
	// AccessionNumbersToSkip is a map of accession numbers to skip during download
	AccessionNumbersToSkip map[string]bool
}

// ToDownload represents a single filing document to be downloaded.
// It contains the necessary URIs and identifiers to locate and save the filing.
type ToDownload struct {
	// RawFilingURI is the URI to the raw filing HTML file
	RawFilingURI string
	// PrimaryDocURI is the URI to the primary document (if any)
	PrimaryDocURI string
	// AccessionNumber is the unique identifier for the filing
	AccessionNumber string
	// DetailsDocSuffix is the suffix for detail documents
	DetailsDocSuffix string
}

// TickerCIKEntry represents a single entry in the ticker to CIK mapping.
// It contains information about a company including its CIK, ticker, name, and exchange.
type TickerCIKEntry struct {
	// CIK is the Central Index Key that identifies the company
	CIK string `json:"cik_str"`
	// Ticker is the stock ticker symbol
	Ticker string `json:"ticker"`
	// Title is the company name
	Title string `json:"title"`
	// Exchange is the stock exchange where the company is listed
	Exchange string `json:"exchange"`
}

// TickerMetadata represents the complete metadata for tickers retrieved from the SEC.
// Format: {"fields":["cik","name","ticker","exchange"],"data":[[320193,"Apple Inc.","AAPL","Nasdaq"],...]}
type TickerMetadata struct {
	// Fields is an array of field names in the data
	Fields []string `json:"fields"`
	// Data is a 2D array containing the actual ticker data
	Data [][]interface{} `json:"data"`
}

// FilingInfo represents information about a specific SEC filing.
// It contains metadata such as accession number, filing date, and form type.
type FilingInfo struct {
	// AccessionNumber is the unique identifier for the filing
	AccessionNumber string `json:"accessionNumber"`
	// FilingDate is the date when the filing was submitted
	FilingDate string `json:"filingDate"`
	// Form is the SEC form type (e.g., "10-K", "8-K")
	Form string `json:"form"`
	// PrimaryDocument is the main document filename in the filing
	PrimaryDocument string `json:"primaryDocument"`
	// Items contains the items covered in the filing (e.g., for 8-K filings)
	Items []string `json:"items"`
}

// SubmissionData represents the complete submission data for a CIK.
// It contains information about all filings made by a company.
type SubmissionData struct {
	// CIK is the Central Index Key that identifies the company
	CIK string `json:"cik"`
	// Filings contains the filing data
	Filings struct {
		// Recent contains the most recent filings
		Recent struct {
			// AccessionNumber is an array of accession numbers
			AccessionNumber []string `json:"accessionNumber"`
			// FilingDate is an array of filing dates
			FilingDate []string `json:"filingDate"`
			// Form is an array of form types
			Form []string `json:"form"`
			// PrimaryDocument is an array of primary document filenames
			PrimaryDocument []string `json:"primaryDocument"`
			// Items is an array of arrays containing items covered in each filing
			Items []string `json:"items"`
		} `json:"recent"`
	} `json:"filings"`
}
