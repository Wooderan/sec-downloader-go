package sec

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSECClient(t *testing.T) {
	companyName := "TestCompany"
	emailAddress := "test@example.com"
	client := NewSECClient(companyName, emailAddress)

	if client == nil {
		t.Errorf("NewSECClient() = nil, want non-nil")
	}

	expectedUserAgent := companyName + " " + emailAddress
	if client.userAgent != expectedUserAgent {
		t.Errorf("NewSECClient() userAgent = %v, want %v", client.userAgent, expectedUserAgent)
	}

	if client.client == nil {
		t.Errorf("NewSECClient() http.Client = nil, want non-nil")
	}

	if client.limiter == nil {
		t.Errorf("NewSECClient() rate.Limiter = nil, want non-nil")
	}
}

func TestGetResponseBody(t *testing.T) {
	tests := []struct {
		name             string
		contentEncoding  string
		compressResponse bool
		wantErr          bool
	}{
		{
			name:             "No compression",
			contentEncoding:  "",
			compressResponse: false,
			wantErr:          false,
		},
		{
			name:             "Gzip compression",
			contentEncoding:  "gzip",
			compressResponse: true,
			wantErr:          false,
		},
		{
			name:             "Gzip header but invalid content",
			contentEncoding:  "gzip",
			compressResponse: false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test response
			responseContent := []byte("test response content")
			var responseBody io.ReadCloser

			if tt.compressResponse {
				// Create gzipped content
				var b bytes.Buffer
				w := gzip.NewWriter(&b)
				_, err := w.Write(responseContent)
				if err != nil {
					t.Fatalf("Failed to gzip content: %v", err)
				}
				w.Close()
				responseBody = io.NopCloser(bytes.NewReader(b.Bytes()))
			} else {
				responseBody = io.NopCloser(bytes.NewReader(responseContent))
			}

			// Create HTTP response
			resp := &http.Response{
				Body:   responseBody,
				Header: make(http.Header),
			}
			if tt.contentEncoding != "" {
				resp.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			// Call function under test
			body, err := getResponseBody(resp)

			// Check for expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("getResponseBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error, verify we can read the content
			if err == nil {
				content, err := io.ReadAll(body)
				if err != nil {
					t.Errorf("Failed to read body: %v", err)
					return
				}

				// If decompression worked correctly, content should match original
				if !tt.wantErr && string(content) != string(responseContent) && tt.compressResponse {
					t.Errorf("getResponseBody() content = %v, want %v", string(content), string(responseContent))
				}

				body.Close()
			}
		})
	}
}

func TestSECClientDownloadFiling(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		wantErr      bool
	}{
		{
			name:         "Successful download",
			statusCode:   http.StatusOK,
			responseBody: "test filing content",
			wantErr:      false,
		},
		{
			name:         "HTTP error",
			statusCode:   http.StatusNotFound,
			responseBody: "not found",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Header.Get("User-Agent") == "" {
					t.Errorf("Request missing User-Agent header")
				}

				// Return response
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client
			client := NewSECClient("TestCompany", "test@example.com")

			// Call function under test
			content, err := client.DownloadFiling(server.URL)

			// Check for expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadFiling() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify response content if no error
			if err == nil && string(content) != tt.responseBody {
				t.Errorf("DownloadFiling() content = %v, want %v", string(content), tt.responseBody)
			}
		})
	}
}

func TestSECClientGetListOfAvailableFilings(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		responseObj *SubmissionData
		wantErr     bool
	}{
		{
			name:       "Successful response",
			statusCode: http.StatusOK,
			responseObj: &SubmissionData{
				CIK: "0000320193",
				Filings: struct {
					Recent struct {
						AccessionNumber []string `json:"accessionNumber"`
						FilingDate      []string `json:"filingDate"`
						Form            []string `json:"form"`
						PrimaryDocument []string `json:"primaryDocument"`
						Items           []string `json:"items"`
					} `json:"recent"`
				}{
					Recent: struct {
						AccessionNumber []string `json:"accessionNumber"`
						FilingDate      []string `json:"filingDate"`
						Form            []string `json:"form"`
						PrimaryDocument []string `json:"primaryDocument"`
						Items           []string `json:"items"`
					}{
						AccessionNumber: []string{"0000320193-22-000001"},
						FilingDate:      []string{"2022-01-01"},
						Form:            []string{"10-K"},
						PrimaryDocument: []string{"primary.htm"},
						Items:           []string{""},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "HTTP error",
			statusCode:  http.StatusNotFound,
			responseObj: nil,
			wantErr:     true,
		},
		{
			name:        "Invalid JSON",
			statusCode:  http.StatusOK,
			responseObj: nil, // Will trigger invalid JSON response
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Header.Get("User-Agent") == "" {
					t.Errorf("Request missing User-Agent header")
				}

				// Return response
				w.WriteHeader(tt.statusCode)

				if tt.responseObj != nil {
					json.NewEncoder(w).Encode(tt.responseObj)
				} else if tt.statusCode == http.StatusOK {
					// Invalid JSON response
					w.Write([]byte("{invalid json}"))
				}
			}))
			defer server.Close()

			// Create client
			client := NewSECClient("TestCompany", "test@example.com")

			// Call function under test
			data, err := client.GetListOfAvailableFilings(server.URL)

			// Check for expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetListOfAvailableFilings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify response data if no error
			if err == nil {
				if data.CIK != tt.responseObj.CIK {
					t.Errorf("GetListOfAvailableFilings() CIK = %v, want %v", data.CIK, tt.responseObj.CIK)
				}

				if len(data.Filings.Recent.AccessionNumber) != len(tt.responseObj.Filings.Recent.AccessionNumber) {
					t.Errorf("GetListOfAvailableFilings() AccessionNumber count = %v, want %v",
						len(data.Filings.Recent.AccessionNumber),
						len(tt.responseObj.Filings.Recent.AccessionNumber))
				}

				if len(data.Filings.Recent.AccessionNumber) > 0 &&
					data.Filings.Recent.AccessionNumber[0] != tt.responseObj.Filings.Recent.AccessionNumber[0] {
					t.Errorf("GetListOfAvailableFilings() AccessionNumber[0] = %v, want %v",
						data.Filings.Recent.AccessionNumber[0],
						tt.responseObj.Filings.Recent.AccessionNumber[0])
				}
			}
		})
	}
}

func TestSECClientGetTickerMetadata(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseObj    *TickerMetadata
		expectedCIKMap map[string]string
		wantErr        bool
	}{
		{
			name:       "Successful response",
			statusCode: http.StatusOK,
			responseObj: &TickerMetadata{
				Fields: []string{"cik", "name", "ticker", "exchange"},
				Data: [][]interface{}{
					{float64(320193), "Apple Inc.", "AAPL", "Nasdaq"},
					{float64(789019), "Microsoft Corporation", "MSFT", "Nasdaq"},
				},
			},
			expectedCIKMap: map[string]string{
				"AAPL": "0000320193",
				"MSFT": "0000789019",
			},
			wantErr: false,
		},
		{
			name:        "HTTP error",
			statusCode:  http.StatusNotFound,
			responseObj: nil,
			wantErr:     true,
		},
		{
			name:        "Invalid JSON",
			statusCode:  http.StatusOK,
			responseObj: nil, // Will trigger invalid JSON response
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Header.Get("User-Agent") == "" {
					t.Errorf("Request missing User-Agent header")
				}

				// Return response
				w.WriteHeader(tt.statusCode)

				if tt.responseObj != nil {
					json.NewEncoder(w).Encode(tt.responseObj)
				} else if tt.statusCode == http.StatusOK {
					// Invalid JSON response
					w.Write([]byte("{invalid json}"))
				}
			}))
			defer server.Close()

			// Create client
			client := NewSECClient("TestCompany", "test@example.com")

			// Test the fetchTickerMetadata directly since we can't override URLCIKMapping
			cikMap, err := client.fetchTickerMetadata(server.URL)

			// Check for expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchTickerMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify response data if no error
			if err == nil {
				if len(cikMap) != len(tt.expectedCIKMap) {
					t.Errorf("fetchTickerMetadata() returned map with %d entries, want %d", len(cikMap), len(tt.expectedCIKMap))
				}

				for ticker, cik := range tt.expectedCIKMap {
					if gotCIK, ok := cikMap[ticker]; !ok || gotCIK != cik {
						t.Errorf("fetchTickerMetadata() for ticker %s = %v, want %v", ticker, gotCIK, cik)
					}
				}
			}
		})
	}
}
