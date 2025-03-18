package sec

import (
	"testing"
	"time"
)

func TestIsCIK(t *testing.T) {
	tests := []struct {
		name        string
		tickerOrCIK string
		want        bool
	}{
		{
			name:        "Valid CIK",
			tickerOrCIK: "0000320193",
			want:        true,
		},
		{
			name:        "Valid CIK without leading zeros",
			tickerOrCIK: "320193",
			want:        true,
		},
		{
			name:        "Ticker symbol",
			tickerOrCIK: "AAPL",
			want:        false,
		},
		{
			name:        "Empty string",
			tickerOrCIK: "",
			want:        false,
		},
		{
			name:        "Mixed alphanumeric",
			tickerOrCIK: "123ABC",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCIK(tt.tickerOrCIK); got != tt.want {
				t.Errorf("IsCIK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndConvertTickerOrCIK(t *testing.T) {
	// Mock ticker to CIK mapping
	tickerToCIKMap := map[string]string{
		"AAPL": "0000320193",
		"MSFT": "0000789019",
		"GOOG": "0001652044",
	}

	tests := []struct {
		name           string
		tickerOrCIK    string
		tickerToCIKMap map[string]string
		want           string
		wantErr        bool
	}{
		{
			name:           "Valid CIK",
			tickerOrCIK:    "320193",
			tickerToCIKMap: tickerToCIKMap,
			want:           "0000320193",
			wantErr:        false,
		},
		{
			name:           "Valid ticker",
			tickerOrCIK:    "AAPL",
			tickerToCIKMap: tickerToCIKMap,
			want:           "0000320193",
			wantErr:        false,
		},
		{
			name:           "Valid ticker lowercase",
			tickerOrCIK:    "msft",
			tickerToCIKMap: tickerToCIKMap,
			want:           "0000789019",
			wantErr:        false,
		},
		{
			name:           "Invalid ticker",
			tickerOrCIK:    "INVALID",
			tickerToCIKMap: tickerToCIKMap,
			want:           "",
			wantErr:        true,
		},
		{
			name:           "Empty string",
			tickerOrCIK:    "",
			tickerToCIKMap: tickerToCIKMap,
			want:           "",
			wantErr:        true,
		},
		{
			name:           "CIK too long",
			tickerOrCIK:    "12345678901234567890",
			tickerToCIKMap: tickerToCIKMap,
			want:           "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndConvertTickerOrCIK(tt.tickerOrCIK, tt.tickerToCIKMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndConvertTickerOrCIK() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateAndConvertTickerOrCIK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndParseDate(t *testing.T) {
	// Reference dates
	validDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		inputDate interface{}
		want      time.Time
		wantErr   bool
	}{
		{
			name:      "Valid time.Time",
			inputDate: validDate,
			want:      validDate,
			wantErr:   false,
		},
		{
			name:      "Valid date string",
			inputDate: "2022-01-01",
			want:      validDate,
			wantErr:   false,
		},
		{
			name:      "Invalid date format",
			inputDate: "01/01/2022",
			want:      time.Time{},
			wantErr:   true,
		},
		{
			name:      "Invalid date string",
			inputDate: "not-a-date",
			want:      time.Time{},
			wantErr:   true,
		},
		{
			name:      "Invalid type",
			inputDate: 123,
			want:      time.Time{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndParseDate(tt.inputDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("ValidateAndParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithinRequestedDateRange(t *testing.T) {
	// Create metadata with date range
	metadata := &DownloadMetadata{
		After:  time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Before: time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name       string
		metadata   *DownloadMetadata
		filingDate string
		want       bool
		wantErr    bool
	}{
		{
			name:       "Date within range",
			metadata:   metadata,
			filingDate: "2022-06-15",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "Date equal to After",
			metadata:   metadata,
			filingDate: "2022-01-01",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "Date equal to Before",
			metadata:   metadata,
			filingDate: "2022-12-31",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "Date before range",
			metadata:   metadata,
			filingDate: "2021-12-31",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "Date after range",
			metadata:   metadata,
			filingDate: "2023-01-01",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "Invalid date format",
			metadata:   metadata,
			filingDate: "invalid-date",
			want:       false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WithinRequestedDateRange(tt.metadata, tt.filingDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithinRequestedDateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WithinRequestedDateRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
