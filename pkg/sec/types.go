package sec

import (
	"time"
)

// DownloadMetadata represents the metadata for a download
type DownloadMetadata struct {
	DownloadFolder         string
	Form                   string
	CIK                    string
	Limit                  int
	After                  time.Time
	Before                 time.Time
	IncludeAmends          bool
	DownloadDetails        bool
	Ticker                 string
	AccessionNumbersToSkip map[string]bool
}

// ToDownload represents a filing to download
type ToDownload struct {
	RawFilingURI     string
	PrimaryDocURI    string
	AccessionNumber  string
	DetailsDocSuffix string
}

// TickerCIKEntry represents an entry in the ticker to CIK mapping
type TickerCIKEntry struct {
	CIK      string `json:"cik_str"`
	Ticker   string `json:"ticker"`
	Title    string `json:"title"`
	Exchange string `json:"exchange"`
}

// TickerMetadata represents the metadata for a ticker
// Format: {"fields":["cik","name","ticker","exchange"],"data":[[320193,"Apple Inc.","AAPL","Nasdaq"],...]}
type TickerMetadata struct {
	Fields []string        `json:"fields"`
	Data   [][]interface{} `json:"data"`
}

// FilingInfo represents information about a filing
type FilingInfo struct {
	AccessionNumber string   `json:"accessionNumber"`
	FilingDate      string   `json:"filingDate"`
	Form            string   `json:"form"`
	PrimaryDocument string   `json:"primaryDocument"`
	Items           []string `json:"items"`
}

// SubmissionData represents the data for a submission
type SubmissionData struct {
	CIK     string `json:"cik"`
	Filings struct {
		Recent struct {
			AccessionNumber []string `json:"accessionNumber"`
			FilingDate      []string `json:"filingDate"`
			Form            []string `json:"form"`
			PrimaryDocument []string `json:"primaryDocument"`
			Items           []string `json:"items"`
		} `json:"recent"`
	} `json:"filings"`
}
