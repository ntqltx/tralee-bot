package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

/*
	this scraper does not currently work
	daft.ie sits behind a cloudflate managed challenge
	that requires a real browser to execute JavaScript
	and obtain a "cf_clearance" cookie before the listings HTML is served.

	I'll try thinking about other ideas how to implement it.
*/

const (
	daftURL   = "https://www.daft.ie/property-for-rent/tralee-kerry"
	daftBase  = "https://www.daft.ie"
	seenPath  = "seen.json"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
)

var nextDataRe = regexp.MustCompile(`<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

func fetchListings() ([]Listing, error) {
	req, err := http.NewRequest("GET", daftURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("daft returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := nextDataRe.FindSubmatch(body)
	if m == nil {
		return nil, fmt.Errorf("__NEXT_DATA__ not found")
	}

	var data NextData
	if err := json.Unmarshal(m[1], &data); err != nil {
		return nil, fmt.Errorf("parse next data: %w", err)
	}

	out := make([]Listing, 0, len(data.Props.PageProps.Listings))
	for _, l := range data.Props.PageProps.Listings {
		if l.Listing.ID == 0 {
			continue
		}
		out = append(out, Listing{
			ID:    l.Listing.ID,
			Title: l.Listing.Title,
			Price: l.Listing.Price,
			URL:   daftBase + l.Listing.SeoFriendlyPath,
		})
	}
	return out, nil
}

func loadSeen() map[int64]bool {
	seen := map[int64]bool{}
	data, err := os.ReadFile(seenPath)

	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("read %s: %v", seenPath, err)
		}
		return seen
	}

	var ids []int64
	if err := json.Unmarshal(data, &ids); err != nil {
		log.Printf("parse %s: %v", seenPath, err)
		return seen
	}
	for _, id := range ids {
		seen[id] = true
	}
	return seen
}

func saveSeen(seen map[int64]bool) {
	ids := make([]int64, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	data, err := json.Marshal(ids)
	if err != nil {
		log.Printf("marshal seen: %v", err)
		return
	}

	if err := os.WriteFile(seenPath, data, 0644); err != nil {
		log.Printf("write %s: %v", seenPath, err)
	}
}

func checkListings() {
	listings, err := fetchListings()
	if err != nil {
		log.Printf("fetch listings: %v", err)
		return
	}
	seen := loadSeen()
	fresh := 0

	for _, l := range listings {
		if seen[l.ID] {
			continue
		}
		seen[l.ID] = true
		fresh++

		msg := fmt.Sprintf("%s\n%s\n%s", l.Title, l.Price, l.URL)
		broadcast(msg)
	}

	if fresh > 0 {
		saveSeen(seen)
		log.Printf("Broadcast %d new listing(s)", fresh)
	} else {
		log.Printf("No new listings (%d total)", len(listings))
	}
}
