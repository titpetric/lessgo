package expression

import (
	"fmt"

	"github.com/titpetric/lessgo/internal/strings"
)

// List represents a LESS list (comma-separated or space-separated values)
type List struct {
	Items []string // Raw string representations of items
}

// NewList creates a new list from items
func NewList(items []string) *List {
	return &List{Items: items}
}

// ParseList parses a comma-separated list
// Handles quoted strings and unquoted values
func ParseList(s string) *List {
	s = strings.TrimSpace(s)
	if s == "" {
		return &List{Items: []string{}}
	}

	var items []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if (ch == '"' || ch == '\'') && !inQuotes {
			// Start quoted string
			inQuotes = true
			quoteChar = ch
			current.WriteByte(ch)
		} else if ch == quoteChar && inQuotes {
			// End quoted string
			inQuotes = false
			current.WriteByte(ch)
		} else if ch == ',' && !inQuotes {
			// End of item
			item := strings.TrimSpace(current.String())
			if item != "" {
				items = append(items, item)
			}
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	// Add last item
	item := strings.TrimSpace(current.String())
	if item != "" {
		items = append(items, item)
	}

	return &List{Items: items}
}

// String returns string representation
func (l *List) String() string {
	return strings.Join(l.Items, ", ")
}

// Length returns the number of items in the list
func (l *List) Length() int {
	return len(l.Items)
}

// Extract returns the item at the given index (1-based)
func (l *List) Extract(index int) (string, error) {
	if index < 1 || index > len(l.Items) {
		return "", fmt.Errorf("index out of range: %d", index)
	}
	return l.Items[index-1], nil
}
