package main

import (
	"log"
	"time"
)

const CHECK_INTERVAL = 15 * time.Minute

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

		broadcast(formatListing(l))
	}

	if fresh > 0 {
		saveSeen(seen)
		log.Printf("Broadcast %d new listing(s)", fresh)
	} else {
		log.Printf("No new listings (%d total)", len(listings))
	}
}

func main() {
	log.Println("Bot started...")
	go handleUpdates()
	checkListings()

	ticker := time.NewTicker(CHECK_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		checkListings()
	}
}
