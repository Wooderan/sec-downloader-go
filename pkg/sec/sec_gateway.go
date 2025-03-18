package sec

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// SECClient represents a client for interacting with the SEC API.
// It handles rate limiting, authentication, and communication with the SEC EDGAR database.
type SECClient struct {
	client    *http.Client
	userAgent string
	limiter   *rate.Limiter
}

// NewSECClient creates a new SEC client with appropriate rate limiting.
//
// Parameters:
//   - companyName: Your company name (required by SEC fair access policy)
//   - emailAddress: Your email address (required by SEC fair access policy)
//
// Returns:
//   - A new SECClient instance configured for SEC API access
//
// Example: NewSECClient("YourCompany", "your@email.com")
func NewSECClient(companyName, emailAddress string) *SECClient {
	userAgent := fmt.Sprintf("%s %s", companyName, emailAddress)

	// 10 requests per second rate limit set by SEC
	limiter := rate.NewLimiter(rate.Limit(SECRequestsPerSecMax), 1)

	return &SECClient{
		client:    &http.Client{Timeout: 30 * time.Second},
		userAgent: userAgent,
		limiter:   limiter,
	}
}

// callSEC makes a rate-limited call to the SEC API.
// It's a helper method that uses the default background context.
func (s *SECClient) callSEC(uri string, host string) (*http.Response, error) {
	return s.callSECWithContext(context.Background(), uri, host)
}

// callSECWithContext makes a rate-limited call to the SEC API with a specific context.
// It respects the SEC's rate limits and sets appropriate headers.
func (s *SECClient) callSECWithContext(ctx context.Context, uri string, host string) (*http.Response, error) {
	// Wait for rate limiter
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", uri, err)
	}

	// Set headers
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Host", host)

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request for %s: %w", uri, err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error %d for %s", resp.StatusCode, uri)
	}

	return resp, nil
}

// getResponseBody handles decompression of gzipped responses.
// It returns a ReadCloser that should be closed by the caller.
func getResponseBody(resp *http.Response) (io.ReadCloser, error) {
	// Check if the response is gzipped
	if resp.Header.Get("Content-Encoding") == "gzip" {
		// Create a gzip reader
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return gzipReader, nil
	}
	return resp.Body, nil
}

// DownloadFiling downloads a filing from the SEC EDGAR database.
//
// Parameters:
//   - uri: The URI of the filing to download
//
// Returns:
//   - The contents of the filing as a byte slice and nil error on success
//   - nil and error on failure
func (s *SECClient) DownloadFiling(uri string) ([]byte, error) {
	// Make the request
	resp, err := s.callSEC(uri, HostWWWSEC)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Get the response body
	body, err := getResponseBody(resp)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// Read the response body
	return io.ReadAll(body)
}

// GetListOfAvailableFilings retrieves the list of available filings for a CIK.
//
// Parameters:
//   - uri: The URI to the submissions file
//
// Returns:
//   - A SubmissionData object containing filing metadata and nil error on success
//   - nil and error on failure
func (s *SECClient) GetListOfAvailableFilings(uri string) (*SubmissionData, error) {
	// Make the request
	resp, err := s.callSEC(uri, HostDataSEC)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Get the response body
	body, err := getResponseBody(resp)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// Decode the JSON
	var submissionData SubmissionData
	if err := json.NewDecoder(body).Decode(&submissionData); err != nil {
		return nil, fmt.Errorf("failed to decode submission data: %w", err)
	}

	return &submissionData, nil
}

// GetTickerMetadata retrieves the ticker to CIK mapping from the SEC.
//
// Returns:
//   - A map of ticker symbols to CIK numbers and nil error on success
//   - nil and error on failure
func (s *SECClient) GetTickerMetadata() (map[string]string, error) {
	// Fetch ticker metadata for all exchanges
	return s.fetchTickerMetadata(URLCIKMapping)
}

// fetchTickerMetadata fetches ticker metadata from a URL.
// It's a helper method that handles the actual API call and JSON processing.
func (s *SECClient) fetchTickerMetadata(url string) (map[string]string, error) {
	// Make the request
	resp, err := s.callSEC(url, HostWWWSEC)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Get the response body
	body, err := getResponseBody(resp)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// Decode the JSON
	var tickerMetadata TickerMetadata
	if err := json.NewDecoder(body).Decode(&tickerMetadata); err != nil {
		return nil, fmt.Errorf("failed to decode ticker metadata: %w", err)
	}

	// Create a map of ticker to CIK
	tickerToCIKMap := make(map[string]string)
	for _, data := range tickerMetadata.Data {
		if len(data) < 3 {
			continue
		}

		// Extract CIK from the data
		cik, ok := data[0].(float64)
		if !ok {
			continue
		}

		// Extract ticker from the data
		ticker, ok := data[2].(string)
		if !ok {
			continue
		}

		// Add to map (convert CIK to string with 10 digits)
		tickerToCIKMap[strings.ToUpper(ticker)] = fmt.Sprintf("%010.0f", cik)
	}

	return tickerToCIKMap, nil
}
