package main

import (
	"log"
	"time"
)

const (
	CHECK_INTERVAL  = 45 * time.Minute
	STARTUP_RETRIES = 12
	STARTUP_BACKOFF = 5 * time.Second
)

func fetchWithRetry(label string) ([]Listing, bool) {
	for i := 0; i < STARTUP_RETRIES; i++ {
		listings, err := fetchListings()
		if err == nil {
			return listings, true
		}
		log.Printf("%s attempt %d/%d: %v", label, i+1, STARTUP_RETRIES, err)
		time.Sleep(STARTUP_BACKOFF)
	}
	return nil, false
}

func checkWithRetry() {
	listings, ok := fetchWithRetry("startup")
	if !ok {
		log.Printf("giving up on startup fetch, will retry at next tick")
		return
	}
	processListings(listings)
}

func seedSeen() {
	log.Println("First run — seeding seen listings (no broadcast)")
	listings, ok := fetchWithRetry("seed")
	if !ok {
		log.Println("seed failed, next tick may broadcast existing listings")
		return
	}
	seen := ids{}
	for _, l := range listings {
		seen[l.ID] = true
	}
	saveSeen(seen)
	log.Printf("Seeded %d listings", len(seen))
}

func processListings(listings []Listing) {
	seen := loadSeen()
	fresh := 0
	for _, l := range listings {
		if seen[l.ID] {
			continue
		}
		seen[l.ID] = true
		fresh++
		broadcast(formatListing(l))
	}
	if fresh > 0 {
		saveSeen(seen)
		log.Printf("Broadcast %d new listing(s)", fresh)
	} else {
		log.Printf("No new listings (%d total)", len(listings))
	}
}

func checkListings() {
	listings, err := fetchListings()
	if err != nil {
		log.Printf("fetch listings: %v", err)
		return
	}
	processListings(listings)
}

func main() {
	log.Println("Bot started...")
	go handleUpdates()

	if len(loadSeen()) == 0 {
		seedSeen()
	} else {
		checkWithRetry()
	}

	ticker := time.NewTicker(CHECK_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		checkListings()
	}
}
