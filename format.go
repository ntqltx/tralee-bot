package main

import (
	"fmt"
	"html"
	"strings"
)

func formatListing(l Listing) string {
	meta := metaLine(l)
	if meta != "" {
		meta = "\n" + html.EscapeString(meta)
	}

	return fmt.Sprintf(
		"🔔<b>New Listing!</b>\n\n%s\n%s%s\n\n<a href=\"%s\">View on Daft.ie</a>",
		html.EscapeString(l.Title),
		html.EscapeString(l.Price),
		meta,
		html.EscapeString(l.URL),
	)
}

func metaLine(l Listing) string {
	parts := make([]string, 0, 3)
	for _, v := range []string{l.PropertyType, l.Bedrooms, l.Bathrooms} {
		if v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, " · ")
}
