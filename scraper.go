package main

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Listing struct {
	ID           int64
	Title        string
	Price        string
	PropertyType string
	Bedrooms     string
	Bathrooms    string
	URL          string
}

type rawListing struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Price           string `json:"price"`
	PropertyType    string `json:"propertyType"`
	NumBedrooms     string `json:"numBedrooms"`
	NumBathrooms   string  `json:"numBathrooms"`
	SeoFriendlyPath string `json:"seoFriendlyPath"`
}

func (r rawListing) toListing() Listing {
	return Listing{
		ID:           r.ID,
		Title:        r.Title,
		Price:        r.Price,
		PropertyType: r.PropertyType,
		Bedrooms:     r.NumBedrooms,
		Bathrooms:    r.NumBathrooms,
		URL:          DAFT_BASE + r.SeoFriendlyPath,
	}
}

type listingEntry struct {
	Listing rawListing `json:"listing"`
}

type pageProps struct {
	Listings []listingEntry `json:"listings"`
}

type nextData struct {
	Props struct {
		PageProps pageProps `json:"pageProps"`
	} `json:"props"`
}

const (
	DAFT_URL  = "https://www.daft.ie/property-for-rent/tralee-kerry"
	DAFT_BASE = "https://www.daft.ie"
)

var nextDataRe = regexp.MustCompile(`<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

func fetchListings() ([]Listing, error) {
	html, err := solve(DAFT_URL)
	if err != nil {
		return nil, err
	}

	m := nextDataRe.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("__NEXT_DATA__ not found")
	}

	var data nextData
	if err := json.Unmarshal([]byte(m[1]), &data); err != nil {
		return nil, fmt.Errorf("parse next data: %w", err)
	}

	entries := data.Props.PageProps.Listings
	out := make([]Listing, 0, len(entries))

	for _, e := range entries {
		if e.Listing.ID == 0 {
			continue
		}
		out = append(out, e.Listing.toListing())
	}
	return out, nil
}
