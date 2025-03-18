package sec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSaveLocation(t *testing.T) {
	tests := []struct {
		name            string
		metadata        *DownloadMetadata
		accessionNumber string
		saveFilename    string
		want            string
	}{
		{
			name: "Using ticker",
			metadata: &DownloadMetadata{
				DownloadFolder: "/test/folder",
				Form:           "10-K",
				CIK:            "0000320193",
				Ticker:         "AAPL",
			},
			accessionNumber: "0000320193-22-000001",
			saveFilename:    "index.html",
			want:            filepath.Join("/test/folder", RootSaveFolderName, "AAPL", "10-K", "0000320193-22-000001", "index.html"),
		},
		{
			name: "Using CIK when ticker not available",
			metadata: &DownloadMetadata{
				DownloadFolder: "/test/folder",
				Form:           "10-K",
				CIK:            "0000320193",
				Ticker:         "",
			},
			accessionNumber: "0000320193-22-000001",
			saveFilename:    "index.html",
			want:            filepath.Join("/test/folder", RootSaveFolderName, "0000320193", "10-K", "0000320193-22-000001", "index.html"),
		},
		{
			name: "Different file and form",
			metadata: &DownloadMetadata{
				DownloadFolder: "/test/folder",
				Form:           "8-K",
				CIK:            "0000789019",
				Ticker:         "MSFT",
			},
			accessionNumber: "0000789019-22-000001",
			saveFilename:    "primary-document.html",
			want:            filepath.Join("/test/folder", RootSaveFolderName, "MSFT", "8-K", "0000789019-22-000001", "primary-document.html"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSaveLocation(tt.metadata, tt.accessionNumber, tt.saveFilename)
			if got != tt.want {
				t.Errorf("GetSaveLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveDocument(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sec-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test data
	testContent := []byte("test document content")
	testPath := filepath.Join(tempDir, "test", "path", "document.html")

	// Save the document
	err = SaveDocument(testContent, testPath)
	if err != nil {
		t.Errorf("SaveDocument() error = %v", err)
		return
	}

	// Verify the document was saved correctly
	savedContent, err := os.ReadFile(testPath)
	if err != nil {
		t.Errorf("Failed to read saved document: %v", err)
		return
	}

	if string(savedContent) != string(testContent) {
		t.Errorf("SaveDocument() saved content = %v, want %v", string(savedContent), string(testContent))
	}

	// Verify directory structure was created
	_, err = os.Stat(filepath.Dir(testPath))
	if err != nil {
		t.Errorf("Failed to create directory structure: %v", err)
	}
}

func TestGetToDownload(t *testing.T) {
	tests := []struct {
		name    string
		cik     string
		accNum  string
		doc     string
		wantErr bool
	}{
		{
			name:    "Valid inputs",
			cik:     "0000320193",
			accNum:  "0000320193-22-000001",
			doc:     "primary.htm",
			wantErr: false,
		},
		{
			name:    "Empty doc",
			cik:     "0000320193",
			accNum:  "0000320193-22-000001",
			doc:     "",
			wantErr: false,
		},
		{
			name:    "Invalid accession number",
			cik:     "0000320193",
			accNum:  "invalid",
			doc:     "primary.htm",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td, err := GetToDownload(tt.cik, tt.accNum, tt.doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetToDownload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify accession number is carried over correctly
				if td.AccessionNumber != tt.accNum {
					t.Errorf("GetToDownload() AccessionNumber = %v, want %v", td.AccessionNumber, tt.accNum)
				}

				// Verify RawFilingURI is set
				if td.RawFilingURI == "" {
					t.Errorf("GetToDownload() RawFilingURI is empty")
				}

				// Verify PrimaryDocURI is set if doc is provided
				if tt.doc != "" && td.PrimaryDocURI == "" {
					t.Errorf("GetToDownload() PrimaryDocURI is empty when doc is provided")
				}

				// Verify PrimaryDocURI is empty if doc is not provided
				if tt.doc == "" && td.PrimaryDocURI != "" {
					t.Errorf("GetToDownload() PrimaryDocURI is not empty when doc is not provided")
				}
			}
		})
	}
}
