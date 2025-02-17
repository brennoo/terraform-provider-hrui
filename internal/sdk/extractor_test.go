package sdk

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractInt(t *testing.T) {
	html := `
	<html>
		<body>
			<p id="number-text"> 42 </p>
			<p id="invalid-text">not a number</p>
			<p id="attr-number" data-value="123"></p>
			<p id="attr-invalid" data-value="abc"></p>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	tests := []struct {
		name       string
		selector   string
		sourceType string
		attr       []string
		expected   int
		shouldFail bool
	}{
		{"Valid text extraction", "#number-text", "text", nil, 42, false},
		{"Invalid text extraction", "#invalid-text", "text", nil, 0, true},
		{"Valid attribute extraction", "#attr-number", "attribute", []string{"data-value"}, 123, false},
		{"Invalid attribute extraction", "#attr-invalid", "attribute", []string{"data-value"}, 0, true},
		{"Missing attribute", "#attr-number", "attribute", nil, 0, true},
		{"Invalid source type", "#number-text", "invalid", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractInt(doc, tt.selector, tt.sourceType, tt.attr...)
			if (err != nil) != tt.shouldFail {
				t.Errorf("unexpected error status: got %v, want failure: %v", err, tt.shouldFail)
			}
			if err == nil && result != tt.expected {
				t.Errorf("unexpected result: got %d, want %d", result, tt.expected)
			}
		})
	}
}
