package sec

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// IsCIK checks if the given string is a CIK (Central Index Key).
// A string is considered to be a CIK if it can be converted to an integer.
//
// Parameters:
//   - tickerOrCIK: The string to check
//
// Returns:
//   - true if the string is a CIK, false otherwise
func IsCIK(tickerOrCIK string) bool {
	_, err := strconv.Atoi(tickerOrCIK)
	return err == nil
}

// ValidateAndConvertTickerOrCIK validates and converts a ticker or CIK to a properly formatted CIK.
// If the input is a ticker, it will be converted to a CIK using the provided mapping.
// If the input is already a CIK, it will be zero-padded to ensure it is exactly 10 digits long.
//
// Parameters:
//   - tickerOrCIK: The ticker symbol or CIK to validate and convert
//   - tickerToCIKMapping: A map of ticker symbols to CIK numbers
//
// Returns:
//   - The properly formatted CIK and nil error on success
//   - Empty string and error on failure
func ValidateAndConvertTickerOrCIK(tickerOrCIK string, tickerToCIKMapping map[string]string) (string, error) {
	tickerOrCIK = strings.TrimSpace(strings.ToUpper(tickerOrCIK))

	// Check for blank tickers or CIKs
	if tickerOrCIK == "" {
		return "", fmt.Errorf("invalid ticker or CIK: please enter a non-blank value")
	}

	// Detect CIKs and ensure that they are properly zero-padded
	if IsCIK(tickerOrCIK) {
		if len(tickerOrCIK) > CIKLength {
			return "", fmt.Errorf("invalid CIK: CIKs must be at most %d digits long", CIKLength)
		}
		// SEC Edgar APIs require zero-padded CIKs, so we must pad CIK with 0s
		// to ensure that it is exactly 10 digits long
		return fmt.Sprintf("%010s", tickerOrCIK), nil
	}

	cik, ok := tickerToCIKMapping[tickerOrCIK]
	if !ok {
		return "", fmt.Errorf("ticker %s is invalid and cannot be mapped to a CIK: please enter a valid ticker or CIK", tickerOrCIK)
	}

	return cik, nil
}

// ValidateAndParseDate validates and parses a date input, which can be either a string or a time.Time object.
// If the input is a string, it must be in the format "YYYY-MM-DD".
//
// Parameters:
//   - inputDate: The date to validate and parse (string in "YYYY-MM-DD" format or time.Time)
//
// Returns:
//   - The parsed time.Time object and nil error on success
//   - Zero time.Time and error on failure
func ValidateAndParseDate(inputDate interface{}) (time.Time, error) {
	switch v := inputDate.(type) {
	case time.Time:
		return v, nil
	case string:
		t, err := time.Parse(DateFormat, v)
		if err != nil {
			return time.Time{}, fmt.Errorf("incorrect date format: please enter a date string of the form YYYY-MM-DD: %w", err)
		}
		return t, nil
	default:
		return time.Time{}, fmt.Errorf("incorrect date input: must be of type string or time.Time")
	}
}

// WithinRequestedDateRange checks if a filing date is within the requested date range.
// The filing date is compared against the After and Before fields in the metadata.
//
// Parameters:
//   - metadata: The download metadata containing the After and Before date bounds
//   - filingDate: The filing date to check as a string in "YYYY-MM-DD" format
//
// Returns:
//   - true and nil error if the date is within range
//   - false and nil error if the date is outside range
//   - false and error if there was an error parsing the filing date
func WithinRequestedDateRange(metadata *DownloadMetadata, filingDate string) (bool, error) {
	targetDate, err := time.Parse(DateFormat, filingDate)
	if err != nil {
		return false, fmt.Errorf("failed to parse filing date: %w", err)
	}

	return !targetDate.Before(metadata.After) && !targetDate.After(metadata.Before), nil
}
