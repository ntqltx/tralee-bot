package main

import (
	"log"
	"time"
)

const (
	CHECK_INTERVAL  = 15 * time.Minute
	STARTUP_RETRIES = 12
	STARTUP_BACKOFF = 5 * time.Second
)

func checkWithRetry() {
	for i := 0; i < STARTUP_RETRIES; i++ {
		listings, err := fetchListings()
		if err == nil {
			processListings(listings)
			return
		}
		log.Printf("startup attempt %d/%d: %v", i+1, STARTUP_RETRIES, err)
		time.Sleep(STARTUP_BACKOFF)
	}
	log.Printf("giving up on startup fetch, will retry at next tick")
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
	checkWithRetry()

	ticker := time.NewTicker(CHECK_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		checkListings()
	}
}
