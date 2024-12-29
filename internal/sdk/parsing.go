package sdk

import (
	"log"
	"strconv"
	"strings"
)

// parseInt parses an integer from a string, supporting optional prefix removal,
// default values, special cases like "Auto" or "Off", and returning nil for special cases if specified.
func parseInt(value string, options ...ParseOption) *int {
	opts := defaultParseOptions()
	for _, option := range options {
		option(&opts)
	}

	if opts.trimPrefix != "" {
		value = strings.TrimPrefix(value, opts.trimPrefix)
	}

	if opts.trimSuffix != "" {
		value = strings.TrimSuffix(value, opts.trimSuffix)
	}

	// Handle special values
	trimmedValue := strings.TrimSpace(value)
	if opts.specialCases[trimmedValue] {
		// If returnNilOnSpecialCases is true, return nil
		if opts.returnNilOnSpecialCases {
			return nil
		}
		// Otherwise, return the default value
		fallbackValue := opts.defaultValue
		return &fallbackValue
	}

	// Attempt to parse the integer
	parsedValue, err := strconv.Atoi(trimmedValue)
	if err != nil {
		if opts.logErrors {
			log.Printf("[DEBUG] Failed to parse int: %s", value)
		}
		parsedValue = opts.defaultValue
	}

	// Apply offset (if any)
	parsedValue += opts.offset

	return &parsedValue
}

// ParseOptions defines options for parsing integers.
type ParseOptions struct {
	trimSuffix              string
	trimPrefix              string
	defaultValue            int
	offset                  int
	logErrors               bool
	specialCases            map[string]bool
	returnNilOnSpecialCases bool
}

// ParseOption modifies ParseOptions.
type ParseOption func(*ParseOptions)

// defaultParseOptions provides default settings for ParseOptions.
func defaultParseOptions() ParseOptions {
	return ParseOptions{
		trimSuffix:              "",
		trimPrefix:              "",
		defaultValue:            0,
		offset:                  0,
		logErrors:               false,
		specialCases:            map[string]bool{},
		returnNilOnSpecialCases: false,
	}
}

// WithTrimSuffix specifies a suffix to trim.
func WithTrimSuffix(suffix string) ParseOption {
	return func(opts *ParseOptions) {
		opts.trimSuffix = suffix
	}
}

// WithTrimPrefix specifies a prefix to trim.
func WithTrimPrefix(prefix string) ParseOption {
	return func(opts *ParseOptions) {
		opts.trimPrefix = prefix
	}
}

// WithDefaultValue specifies a default value to return on error.
func WithDefaultValue(value int) ParseOption {
	return func(opts *ParseOptions) {
		opts.defaultValue = value
	}
}

// WithOffset specifies an offset to apply to parsed values.
func WithOffset(offset int) ParseOption {
	return func(opts *ParseOptions) {
		opts.offset = offset
	}
}

// WithLogging enables error logging.
func WithLogging() ParseOption {
	return func(opts *ParseOptions) {
		opts.logErrors = true
	}
}

// WithSpecialCases defines special-case strings that map to the default value.
func WithSpecialCases(cases ...string) ParseOption {
	return func(opts *ParseOptions) {
		for _, c := range cases {
			opts.specialCases[c] = true
		}
	}
}

// WithReturnNilOnSpecialCases defines that special cases should return nil instead of the default value.
func WithReturnNilOnSpecialCases() ParseOption {
	return func(opts *ParseOptions) {
		opts.returnNilOnSpecialCases = true
	}
}
