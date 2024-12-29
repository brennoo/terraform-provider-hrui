package sdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// extractInt extracts an integer either from the text content or an attribute of a given selector.
func extractInt(doc *goquery.Document, selector string, sourceType string, attr ...string) (int, error) {
	var value string

	// Determine the source of the value
	switch sourceType {
	case "text":
		text, err := extractText(doc, selector)
		if err != nil {
			return 0, err
		}
		value = text
	case "attribute":
		if len(attr) == 0 {
			return 0, fmt.Errorf("attribute name must be provided for sourceType 'attribute'")
		}
		value = doc.Find(selector).AttrOr(attr[0], "")
	default:
		return 0, fmt.Errorf("invalid sourceType: %s", sourceType)
	}

	// Parse the integer value
	return strconv.Atoi(strings.TrimSpace(value))
}
