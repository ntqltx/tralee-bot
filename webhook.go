package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const maxWebhookBodyBytes = 1 << 20 // 1 MiB is plenty for an RSS batch.

var processMu sync.Mutex

type webhookListing struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type webhookResponse struct {
	Received    int `json:"received"`
	Broadcasted int `json:"broadcasted"`
}

func newListingsWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)
	defer r.Body.Close()

	var payload []webhookListing
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON payload: %v", err))
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload: multiple JSON values")
		return
	}

	listings, err := normalizeWebhookListings(payload)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	broadcasted := processWebhookListings(listings)
	log.Printf("Webhook received %d listing(s), broadcast %d new listing(s)", len(listings), broadcasted)

	writeJSON(w, http.StatusOK, webhookResponse{
		Received:    len(listings),
		Broadcasted: broadcasted,
	})
}

func normalizeWebhookListings(payload []webhookListing) ([]Listing, error) {
	out := make([]Listing, 0, len(payload))
	for i, item := range payload {
		item.Title = strings.TrimSpace(item.Title)
		item.URL = strings.TrimSpace(item.URL)

		if item.ID <= 0 {
			return nil, fmt.Errorf("listing at index %d has invalid id", i)
		}
		if item.Title == "" {
			return nil, fmt.Errorf("listing at index %d has empty title", i)
		}
		if item.URL == "" {
			return nil, fmt.Errorf("listing at index %d has empty url", i)
		}
		parsed, err := url.ParseRequestURI(item.URL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("listing at index %d has invalid url", i)
		}

		out = append(out, Listing{
			ID:    item.ID,
			Title: item.Title,
			URL:   item.URL,
		})
	}
	return out, nil
}

func processWebhookListings(listings []Listing) int {
	processMu.Lock()

	// Production dedupe lives here. This repo currently persists seen listing IDs
	// to JSON on Railway's /data volume; this can be swapped for Postgres/Redis
	// without changing the webhook contract.
	seen := loadSeen()
	fresh := make([]Listing, 0, len(listings))
	for _, l := range listings {
		if seen[l.ID] {
			continue
		}
		seen[l.ID] = true
		fresh = append(fresh, l)
	}
	if len(fresh) > 0 {
		saveSeen(seen)
	}
	processMu.Unlock()

	for _, l := range fresh {
		broadcast(formatListing(l))
	}
	return len(fresh)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}
