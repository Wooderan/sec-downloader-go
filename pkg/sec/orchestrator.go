package sec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetSaveLocation returns the path where a filing should be saved
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

// SaveDocument saves a document to disk
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

// AggregateFilingsToDownload aggregates the filings to download
func AggregateFilingsToDownload(metadata *DownloadMetadata, client *SECClient) ([]ToDownload, error) {
	// Format the submission URL
	submissionFile := fmt.Sprintf(SubmissionFileFormat, metadata.CIK)
	submissionURL := fmt.Sprintf(URLSubmissions, submissionFile)

	// Get the list of available filings
	submissionData, err := client.GetListOfAvailableFilings(submissionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of available filings: %w", err)
	}

	// Prepare the list of filings to download
	var toDownload []ToDownload
	filingCount := 0

	// Get the recent filings
	recent := submissionData.Filings.Recent

	// Iterate through the filings
	for i := 0; i < len(recent.AccessionNumber); i++ {
		// Check if we've reached the limit
		if filingCount >= metadata.Limit {
			break
		}

		// Get the filing form
		form := recent.Form[i]

		// Check if the form matches what we're looking for
		if !strings.EqualFold(form, metadata.Form) {
			// If we're not including amends, skip this filing
			if !metadata.IncludeAmends || !strings.HasSuffix(strings.ToUpper(form), AmendsSuffix) {
				continue
			}

			// If we are including amends, check if the base form matches
			baseForm := strings.TrimSuffix(strings.ToUpper(form), AmendsSuffix)
			if !strings.EqualFold(baseForm, metadata.Form) {
				continue
			}
		}

		// Check if the filing date is within the requested range
		withinRange, err := WithinRequestedDateRange(metadata, recent.FilingDate[i])
		if err != nil {
			return nil, fmt.Errorf("failed to check if filing date is within range: %w", err)
		}

		if !withinRange {
			continue
		}

		// Check if we should skip this accession number
		accNum := recent.AccessionNumber[i]
		if metadata.AccessionNumbersToSkip != nil && metadata.AccessionNumbersToSkip[accNum] {
			continue
		}

		// Get the primary document
		primaryDoc := recent.PrimaryDocument[i]

		// Create the ToDownload object
		td, err := GetToDownload(metadata.CIK, accNum, primaryDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to get download info: %w", err)
		}

		toDownload = append(toDownload, *td)
		filingCount++
	}

	return toDownload, nil
}

// GetToDownload creates a ToDownload object for a filing
func GetToDownload(cik, accNum, doc string) (*ToDownload, error) {
	// Format the accession number for the URL (remove dashes)
	accNumNoDash := strings.ReplaceAll(accNum, "-", "")

	// Create the raw filing URI (index page that contains links to all documents)
	rawFilingURI := fmt.Sprintf(URLFilingArchive, cik, accNumNoDash, accNum)

	// Create the primary document URI
	var primaryDocURI string
	if doc != "" {
		primaryDocURI = fmt.Sprintf(URLFiling, cik, accNumNoDash, doc)
	}

	// Get the file extension for the details document
	var detailsDocSuffix string
	if doc != "" {
		ext := filepath.Ext(doc)
		if ext != "" {
			detailsDocSuffix = ext
		}
	}

	return &ToDownload{
		RawFilingURI:     rawFilingURI,
		PrimaryDocURI:    primaryDocURI,
		AccessionNumber:  accNum,
		DetailsDocSuffix: detailsDocSuffix,
	}, nil
}

// FetchAndSaveFilings fetches and saves filings
func FetchAndSaveFilings(metadata *DownloadMetadata, client *SECClient) (int, error) {
	// Get the list of filings to download
	toDownload, err := AggregateFilingsToDownload(metadata, client)
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate filings to download: %w", err)
	}

	// Download and save each filing
	for _, td := range toDownload {
		// Download the raw filing
		rawFiling, err := client.DownloadFiling(td.RawFilingURI)
		if err != nil {
			return 0, fmt.Errorf("failed to download raw filing for accession number %s: %w", td.AccessionNumber, err)
		}

		// Save the raw filing
		rawFilingPath := GetSaveLocation(metadata, td.AccessionNumber, FilingFullSubmissionFilename)
		if err := SaveDocument(rawFiling, rawFilingPath); err != nil {
			return 0, fmt.Errorf("failed to save raw filing for accession number %s to %s: %w", td.AccessionNumber, rawFilingPath, err)
		}

		// If the primary document exists, download and save it
		if td.PrimaryDocURI != "" {
			// Download the primary document
			primaryDoc, err := client.DownloadFiling(td.PrimaryDocURI)
			if err != nil {
				return 0, fmt.Errorf("failed to download primary document for accession number %s: %w", td.AccessionNumber, err)
			}

			// Save the primary document
			primaryDocFilename := PrimaryDocFilenameStem
			if td.DetailsDocSuffix != "" {
				primaryDocFilename += td.DetailsDocSuffix
			}
			primaryDocPath := GetSaveLocation(metadata, td.AccessionNumber, primaryDocFilename)
			if err := SaveDocument(primaryDoc, primaryDocPath); err != nil {
				return 0, fmt.Errorf("failed to save primary document for accession number %s to %s: %w", td.AccessionNumber, primaryDocPath, err)
			}
		}
	}

	return len(toDownload), nil
}
