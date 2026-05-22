package main

import (
	"encoding/json"
	"log"
	"os"
)

const DEFAULT_PATH string = "seen.json"

func loadSeen() map[int64]bool {
	seen := map[int64]bool{}
	data, err := os.ReadFile(DEFAULT_PATH)

	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("read %s: %v", DEFAULT_PATH, err)
		}
		return seen
	}

	var ids []int64
	if err := json.Unmarshal(data, &ids); err != nil {
		log.Printf("parse %s: %v", DEFAULT_PATH, err)
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
	if err := os.WriteFile(DEFAULT_PATH, data, 0644); err != nil {
		log.Printf("write %s: %v", DEFAULT_PATH, err)
	}
}
