package main

import (
	"log"
	"time"
)

const (
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

func main() {
	log.Println("Single-shot run started")

	subscribers, err := kvListSubscribers()
	if err != nil {
		log.Fatalf("load subscribers: %v", err)
	}
	log.Printf("Loaded %d subscriber(s)", len(subscribers))

	oldSeen, err := kvGetSeen()
	if err != nil {
		log.Fatalf("load seen: %v", err)
	}
	log.Printf("Loaded %d seen listing(s)", len(oldSeen))

	firstRun := len(oldSeen) == 0

	listings, ok := fetchWithRetry("fetch")
	if !ok {
		log.Fatal("giving up: flaresolverr never returned a successful response")
	}

	currentSeen := make(map[int64]bool, len(listings))
	fresh := make([]Listing, 0, len(listings))
	for _, l := range listings {
		currentSeen[l.ID] = true
		if !oldSeen[l.ID] {
			fresh = append(fresh, l)
		}
	}
	pruned := 0
	for id := range oldSeen {
		if !currentSeen[id] {
			pruned++
		}
	}

	if firstRun {
		log.Printf("First run - seeding %d listing(s) without broadcasting", len(fresh))
	} else {
		log.Printf("Broadcasting %d new listing(s) to %d subscriber(s)", len(fresh), len(subscribers))
		for _, l := range fresh {
			broadcast(subscribers, formatListing(l))
		}
	}

	if len(fresh) > 0 || pruned > 0 {
		if err := kvPutSeen(currentSeen); err != nil {
			log.Fatalf("save seen: %v", err)
		}
		log.Printf("Persisted %d seen listing ID(s) to KV (%d new, %d pruned)", len(currentSeen), len(fresh), pruned)
	} else {
		log.Println("No changes, KV unchanged")
	}

	log.Println("Done")
}
