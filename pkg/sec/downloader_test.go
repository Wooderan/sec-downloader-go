package sec

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestWithLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{
			name:  "Positive limit",
			limit: 10,
			want:  10,
		},
		{
			name:  "Zero limit should set max",
			limit: 0,
			want:  math.MaxInt32,
		},
		{
			name:  "Negative limit should set max",
			limit: -5,
			want:  math.MaxInt32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &DownloadMetadata{}
			option := WithLimit(tt.limit)
			option(metadata)

			if metadata.Limit != tt.want {
				t.Errorf("WithLimit() set limit to %v, want %v", metadata.Limit, tt.want)
			}
		})
	}
}

func TestWithDateRange(t *testing.T) {
	// Reference dates
	testDate1 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	testDate2 := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		after      interface{}
		before     interface{}
		wantAfter  time.Time
		wantBefore time.Time
	}{
		{
			name:       "Both time.Time",
			after:      testDate1,
			before:     testDate2,
			wantAfter:  testDate1,
			wantBefore: testDate2,
		},
		{
			name:       "Both strings",
			after:      "2022-01-01",
			before:     "2023-12-31",
			wantAfter:  testDate1,
			wantBefore: testDate2,
		},
		{
			name:       "After time.Time, Before string",
			after:      testDate1,
			before:     "2023-12-31",
			wantAfter:  testDate1,
			wantBefore: testDate2,
		},
		{
			name:       "After string, Before time.Time",
			after:      "2022-01-01",
			before:     testDate2,
			wantAfter:  testDate1,
			wantBefore: testDate2,
		},
		{
			name:       "Invalid after string, keep default",
			after:      "invalid-date",
			before:     testDate2,
			wantAfter:  time.Time{}, // Will use default in metadata
			wantBefore: testDate2,
		},
		{
			name:       "Invalid before string, keep default",
			after:      testDate1,
			before:     "invalid-date",
			wantAfter:  testDate1,
			wantBefore: time.Time{}, // Will use default in metadata
		},
		{
			name:       "Both nil, keep defaults",
			after:      nil,
			before:     nil,
			wantAfter:  time.Time{}, // Will use default in metadata
			wantBefore: time.Time{}, // Will use default in metadata
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &DownloadMetadata{}
			option := WithDateRange(tt.after, tt.before)
			option(metadata)

			// Skip checking dates if they should be default (zero time)
			if !tt.wantAfter.IsZero() && !metadata.After.Equal(tt.wantAfter) {
				t.Errorf("WithDateRange() set After to %v, want %v", metadata.After, tt.wantAfter)
			}

			if !tt.wantBefore.IsZero() && !metadata.Before.Equal(tt.wantBefore) {
				t.Errorf("WithDateRange() set Before to %v, want %v", metadata.Before, tt.wantBefore)
			}
		})
	}
}

func TestWithIncludeAmends(t *testing.T) {
	tests := []struct {
		name          string
		includeAmends bool
		want          bool
	}{
		{
			name:          "Include amendments true",
			includeAmends: true,
			want:          true,
		},
		{
			name:          "Include amendments false",
			includeAmends: false,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &DownloadMetadata{}
			option := WithIncludeAmends(tt.includeAmends)
			option(metadata)

			if metadata.IncludeAmends != tt.want {
				t.Errorf("WithIncludeAmends() set IncludeAmends to %v, want %v", metadata.IncludeAmends, tt.want)
			}
		})
	}
}

func TestWithDownloadDetails(t *testing.T) {
	tests := []struct {
		name            string
		downloadDetails bool
		want            bool
	}{
		{
			name:            "Download details true",
			downloadDetails: true,
			want:            true,
		},
		{
			name:            "Download details false",
			downloadDetails: false,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &DownloadMetadata{}
			option := WithDownloadDetails(tt.downloadDetails)
			option(metadata)

			if metadata.DownloadDetails != tt.want {
				t.Errorf("WithDownloadDetails() set DownloadDetails to %v, want %v", metadata.DownloadDetails, tt.want)
			}
		})
	}
}

func TestWithAccessionNumbersToSkip(t *testing.T) {
	tests := []struct {
		name                   string
		accessionNumbersToSkip map[string]bool
		want                   map[string]bool
	}{
		{
			name: "With accession numbers",
			accessionNumbersToSkip: map[string]bool{
				"0001193125-22-067124": true,
				"0001193125-22-123456": true,
			},
			want: map[string]bool{
				"0001193125-22-067124": true,
				"0001193125-22-123456": true,
			},
		},
		{
			name:                   "Empty map",
			accessionNumbersToSkip: map[string]bool{},
			want:                   map[string]bool{},
		},
		{
			name:                   "Nil map",
			accessionNumbersToSkip: nil,
			want:                   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &DownloadMetadata{}
			option := WithAccessionNumbersToSkip(tt.accessionNumbersToSkip)
			option(metadata)

			if !reflect.DeepEqual(metadata.AccessionNumbersToSkip, tt.want) {
				t.Errorf("WithAccessionNumbersToSkip() set AccessionNumbersToSkip to %v, want %v", metadata.AccessionNumbersToSkip, tt.want)
			}
		})
	}
}

// Mock SEC client for testing the Downloader
type MockSECClient struct {
	tickerToCIKMap map[string]string
}

func (m *MockSECClient) GetTickerMetadata() (map[string]string, error) {
	return m.tickerToCIKMap, nil
}

func TestDownloaderGet(t *testing.T) {
	// We'll test that Get properly creates the options and calls GetWithOptions
	// This can be a partial test since we can't easily mock the full call chain

	// Setup a downloader with mock client
	downloader := &Downloader{
		client:         nil, // Not needed for this test
		downloadFolder: "/test/folder",
		tickerToCIKMap: map[string]string{
			"AAPL": "0000320193",
		},
	}

	// Since we can't easily mock FetchAndSaveFilings, we just want to verify that
	// the Get method properly creates the options and passes them to GetWithOptions.
	// We can do this by testing the input parameter handling.

	// Test basic input handling
	testCases := []struct {
		name                   string
		form                   string
		tickerOrCIK            string
		limit                  int
		after                  interface{}
		before                 interface{}
		includeAmends          bool
		downloadDetails        bool
		accessionNumbersToSkip map[string]bool
		wantErr                bool
	}{
		{
			name:        "Unsupported form",
			form:        "UNSUPPORTED",
			tickerOrCIK: "AAPL",
			limit:       1,
			wantErr:     true,
		},
		{
			name:        "Invalid ticker",
			form:        "10-K",
			tickerOrCIK: "INVALID",
			limit:       1,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := downloader.Get(
				tc.form,
				tc.tickerOrCIK,
				tc.limit,
				tc.after,
				tc.before,
				tc.includeAmends,
				tc.downloadDetails,
				tc.accessionNumbersToSkip,
			)

			if (err != nil) != tc.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
