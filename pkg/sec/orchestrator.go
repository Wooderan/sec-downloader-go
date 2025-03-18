package sec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetSaveLocation returns the path where a filing should be saved.
// It constructs a directory path based on the company identifier, form type, and accession number.
//
// Parameters:
//   - metadata: The download metadata containing configuration options
//   - accessionNumber: The accession number of the filing
//   - saveFilename: The filename to save the document as
//
// Returns:
//   - A string containing the full path where the filing should be saved
func GetSaveLocation(metadata *DownloadMetadata, accessionNumber, saveFilename string) string {
	companyIdentifier := metadata.Ticker
	if companyIdentifier == "" {
		companyIdentifier = metadata.CIK
	}

	return filepath.Join(
		metadata.DownloadFolder,
		RootSaveFolderName,
		companyIdentifier,
		metadata.Form,
		accessionNumber,
		saveFilename,
	)
}

// SaveDocument saves a document to disk, creating any necessary directories.
//
// Parameters:
//   - filingContents: The raw content of the filing as a byte slice
//   - savePath: The full path where the filing should be saved
//
// Returns:
//   - nil on success, error if the save operation fails
func SaveDocument(filingContents []byte, savePath string) error {
	// Create all parent directories as needed
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write content to file
	if err := os.WriteFile(savePath, filingContents, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// AggregateFilingsToDownload aggregates the filings to download based on download metadata.
// It fetches the filing list from the SEC and filters it according to the specified criteria.
//
// Parameters:
//   - metadata: The download metadata containing filtering options
//   - client: The SEC client to use for API requests
//
// Returns:
//   - A slice of ToDownload objects and nil error on success
//   - nil and error on failure
func AggregateFilingsToDownload(metadata *DownloadMetadata, client *SECClient) ([]ToDownload, error) {
	// Format the submission URL
	submissionFile := fmt.Sprintf(SubmissionFileFormat, metadata.CIK)
	submissionURL := fmt.Sprintf(URLSubmissions, submissionFile)

	// Get the list of available filings
	submissionData, err := client.GetListOfAvailableFilings(submissionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of available filings: %w", err)
	}

	// Filter the filings based on the metadata
	filings := submissionData.Filings.Recent
	var toDownload []ToDownload
	for i := 0; i < len(filings.AccessionNumber) && len(toDownload) < metadata.Limit; i++ {
		// Get the form for this filing
		form := filings.Form[i]

		// Skip if form doesn't match and we're not checking for amendments
		if !strings.EqualFold(form, metadata.Form) {
			if !metadata.IncludeAmends || !strings.EqualFold(form, metadata.Form+AmendsSuffix) {
				continue
			}
		}

		// Parse the filing date
		filingDateStr := filings.FilingDate[i]
		filingDate, err := time.Parse(DateFormat, filingDateStr)
		if err != nil {
			// Skip filings with invalid dates
			continue
		}

		// Skip filings outside the date range
		if filingDate.Before(metadata.After) || filingDate.After(metadata.Before) {
			continue
		}

		// Skip filings with specified accession numbers
		accessionNumber := filings.AccessionNumber[i]
		if metadata.AccessionNumbersToSkip != nil && metadata.AccessionNumbersToSkip[accessionNumber] {
			continue
		}

		// Get the document to download
		doc := filings.PrimaryDocument[i]
		td, err := GetToDownload(metadata.CIK, accessionNumber, doc)
		if err != nil {
			return nil, fmt.Errorf("failed to get download URL for accession number %s: %w", accessionNumber, err)
		}

		// Add to the list
		toDownload = append(toDownload, *td)
	}

	return toDownload, nil
}

// GetToDownload constructs a ToDownload object with the appropriate URLs for a filing.
//
// Parameters:
//   - cik: The Central Index Key of the company
//   - accNum: The accession number of the filing
//   - doc: The primary document filename
//
// Returns:
//   - A ToDownload object and nil error on success
//   - nil and error on failure
func GetToDownload(cik, accNum, doc string) (*ToDownload, error) {
	// Remove dashes from accession number
	rawAccNum := strings.ReplaceAll(accNum, "-", "")

	// Calculate the URLs for the filing
	if len(rawAccNum) != 18 {
		return nil, fmt.Errorf("invalid accession number: %s", accNum)
	}

	// Calculate the base URL and archive URLs
	rawFilingURL := fmt.Sprintf(URLFilingArchive, cik, rawAccNum, accNum)

	// Determine the primary document URI if available
	var primaryDocURI string
	if doc != "" {
		primaryDocURI = fmt.Sprintf(URLFiling, cik, rawAccNum, doc)
	}

	// Determine the details document suffix (e.g., for form 4 XML)
	detailsDocSuffix := ""
	if strings.EqualFold(doc, "primary-document.html") {
		detailsDocSuffix = "-index-headers.html"
	} else if strings.HasSuffix(strings.ToLower(doc), ".htm") || strings.HasSuffix(strings.ToLower(doc), ".html") {
		detailsDocSuffix = strings.TrimSuffix(strings.TrimSuffix(doc, ".htm"), ".html") + "-index-headers.html"
	}

	return &ToDownload{
		RawFilingURI:     rawFilingURL,
		PrimaryDocURI:    primaryDocURI,
		AccessionNumber:  accNum,
		DetailsDocSuffix: detailsDocSuffix,
	}, nil
}

// FetchAndSaveFilings fetches and saves filings based on the download metadata.
// This is the main orchestration function that ties together the entire download process.
//
// Parameters:
//   - metadata: The download metadata containing configuration options
//   - client: The SEC client to use for API requests
//
// Returns:
//   - The number of filings downloaded and nil error on success
//   - 0 and error on failure
func FetchAndSaveFilings(metadata *DownloadMetadata, client *SECClient) (int, error) {
	// Get the list of filings to download
	toDownload, err := AggregateFilingsToDownload(metadata, client)
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate filings to download: %w", err)
	}

	// Download and save each filing
	downloadCount := 0
	for _, td := range toDownload {
		// Download index.html
		indexContents, err := client.DownloadFiling(td.RawFilingURI)
		if err != nil {
			continue
		}

		// Save index.html
		savePath := GetSaveLocation(metadata, td.AccessionNumber, FilingFullSubmissionFilename)
		if err := SaveDocument(indexContents, savePath); err != nil {
			continue
		}

		// Download primary document if available
		if td.PrimaryDocURI != "" {
			primaryContents, err := client.DownloadFiling(td.PrimaryDocURI)
			if err == nil {
				// Extract filename from primary document URI
				_, primaryFileName := filepath.Split(td.PrimaryDocURI)
				primarySavePath := GetSaveLocation(metadata, td.AccessionNumber, primaryFileName)
				_ = SaveDocument(primaryContents, primarySavePath)
			}
		}

		// Download details document if requested
		if metadata.DownloadDetails && td.DetailsDocSuffix != "" {
			// Calculate the details URL
			rawAccNum := strings.ReplaceAll(td.AccessionNumber, "-", "")
			detailsURL := fmt.Sprintf(URLFiling, metadata.CIK, rawAccNum, rawAccNum+td.DetailsDocSuffix)

			// Download details document
			detailsContents, err := client.DownloadFiling(detailsURL)
			if err == nil {
				detailsSavePath := GetSaveLocation(metadata, td.AccessionNumber, fmt.Sprintf("index%s", td.DetailsDocSuffix))
				_ = SaveDocument(detailsContents, detailsSavePath)
			}
		}

		downloadCount++
	}

	return downloadCount, nil
}
