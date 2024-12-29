package sdk

import (
	"testing"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		options      []ParseOption
		expected     *int
		shouldBeNil  bool
	}{
		{
			name:     "Valid integer without options",
			value:    "42",
			options:  nil,
			expected: intPtr(42),
		},
		{
			name:     "TrimPrefix applied",
			value:    "v42",
			options:  []ParseOption{WithTrimPrefix("v")},
			expected: intPtr(42),
		},
		{
			name:     "TrimSuffix applied",
			value:    "42ms",
			options:  []ParseOption{WithTrimSuffix("ms")},
			expected: intPtr(42),
		},
		{
			name:     "Default value on parsing error",
			value:    "invalid",
			options:  []ParseOption{WithDefaultValue(99)},
			expected: intPtr(99),
		},
		{
			name:     "Special case returns default",
			value:    "Auto",
			options:  []ParseOption{WithSpecialCases("Auto"), WithDefaultValue(5)},
			expected: intPtr(5),
		},
		{
			name:        "Special case returns nil",
			value:       "Auto",
			options:     []ParseOption{WithSpecialCases("Auto"), WithReturnNilOnSpecialCases()},
			expected:    nil,
			shouldBeNil: true,
		},
		{
			name:     "Offset applied",
			value:    "10",
			options:  []ParseOption{WithOffset(5)},
			expected: intPtr(15),
		},
		{
			name:     "Combination of prefix, suffix, and offset",
			value:    "abc42ms",
			options:  []ParseOption{WithTrimPrefix("abc"), WithTrimSuffix("ms"), WithOffset(8)},
			expected: intPtr(50),
		},
		{
			name:     "Logging enabled with invalid value",
			value:    "invalid",
			options:  []ParseOption{WithLogging(), WithDefaultValue(0)},
			expected: intPtr(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInt(tt.value, tt.options...)

			if tt.shouldBeNil {
				if result != nil {
					t.Errorf("Expected nil but got %v", *result)
				}
				return
			}

			if result == nil || *result != *tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

// Helper to return a pointer to an int
func intPtr(i int) *int {
	return &i
}

