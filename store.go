package main

import (
	"encoding/json"
	"log"
	"os"
)

const DEFAULT_PATH string = "seen.json"

func seenPath() string {
	if p := os.Getenv("SEEN_PATH"); p != "" {
		return p
	}
	return DEFAULT_PATH
}

func loadSeen() map[int64]bool {
	seen := map[int64]bool{}
	path := seenPath()
	data, err := os.ReadFile(path)

	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("read %s: %v", path, err)
		}
		return seen
	}

	var ids []int64
	if err := json.Unmarshal(data, &ids); err != nil {
		log.Printf("parse %s: %v", path, err)
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
	path := seenPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("write %s: %v", path, err)
	}
}
