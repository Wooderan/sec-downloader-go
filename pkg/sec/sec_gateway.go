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

// SECClient represents a client for interacting with the SEC API
type SECClient struct {
	client    *http.Client
	userAgent string
	limiter   *rate.Limiter
}

// NewSECClient creates a new SEC client
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

// callSEC makes a rate-limited call to the SEC API
func (s *SECClient) callSEC(uri string, host string) (*http.Response, error) {
	return s.callSECWithContext(context.Background(), uri, host)
}

// callSECWithContext makes a rate-limited call to the SEC API with context
func (s *SECClient) callSECWithContext(ctx context.Context, uri string, host string) (*http.Response, error) {
	// Wait for rate limiter
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", uri, err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Host", host)
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to %s: %w", uri, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("SEC API returned non-200 status code: %d for URL: %s", resp.StatusCode, uri)
	}

	return resp, nil
}

// getResponseBody reads the response body, handling gzip compression if necessary
func getResponseBody(resp *http.Response) (io.ReadCloser, error) {
	// Check if the response is gzip-encoded
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		// Create a gzip reader
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return gzReader, nil
	}

	return resp.Body, nil
}

// DownloadFiling downloads a filing from the SEC
func (s *SECClient) DownloadFiling(uri string) ([]byte, error) {
	resp, err := s.callSEC(uri, HostWWWSEC)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader, err := getResponseBody(resp)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// GetListOfAvailableFilings gets a list of available filings for a CIK
func (s *SECClient) GetListOfAvailableFilings(uri string) (*SubmissionData, error) {
	resp, err := s.callSEC(uri, HostDataSEC)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader, err := getResponseBody(resp)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var data SubmissionData
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode submission data: %w", err)
	}

	return &data, nil
}

// GetTickerMetadata gets the ticker to CIK mapping
func (s *SECClient) GetTickerMetadata() (map[string]string, error) {
	// Try the primary URL first
	tickerToCIKMap, err := s.fetchTickerMetadata(URLCIKMapping)
	if err != nil || len(tickerToCIKMap) == 0 {
		return nil, fmt.Errorf("failed to get ticker metadata : %w", err)
	}

	return tickerToCIKMap, nil
}

// fetchTickerMetadata fetches ticker metadata from a specific URL
func (s *SECClient) fetchTickerMetadata(url string) (map[string]string, error) {
	resp, err := s.callSEC(url, HostWWWSEC)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader, err := getResponseBody(resp)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Read the entire response body
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Decode the JSON response
	var metadata TickerMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to decode ticker metadata: %w", err)
	}

	// Find the indices for CIK and ticker in the fields array
	var cikIndex, tickerIndex int = -1, -1
	for i, field := range metadata.Fields {
		switch strings.ToLower(field) {
		case "cik", "cik_str":
			cikIndex = i
		case "ticker":
			tickerIndex = i
		}
	}

	if cikIndex == -1 || tickerIndex == -1 {
		return nil, fmt.Errorf("required fields 'cik' and 'ticker' not found in response")
	}

	// Create the ticker to CIK map
	tickerToCIKMap := make(map[string]string)
	for _, entry := range metadata.Data {
		if len(entry) <= cikIndex || len(entry) <= tickerIndex {
			continue
		}

		// Extract CIK (could be float64, int, or string)
		var cikStr string
		switch cik := entry[cikIndex].(type) {
		case float64:
			cikStr = fmt.Sprintf("%d", int(cik))
		case int:
			cikStr = fmt.Sprintf("%d", cik)
		case string:
			cikStr = cik
		default:
			continue
		}

		// Extract ticker (should be string)
		ticker, ok := entry[tickerIndex].(string)
		if !ok {
			continue
		}

		// Pad the CIK with leading zeros to ensure it's 10 digits
		paddedCIK := fmt.Sprintf("%010s", cikStr)
		tickerToCIKMap[ticker] = paddedCIK
	}

	if len(tickerToCIKMap) == 0 {
		return nil, fmt.Errorf("no ticker to CIK mappings found in the response")
	}

	return tickerToCIKMap, nil
}
