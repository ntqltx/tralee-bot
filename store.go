package main

import (
	"encoding/json"
	"log"
	"os"
)

type ids = map[int64]bool

const (
	DEFAULT_SEEN_PATH = "seen.json"
	DEFAULT_SUBS_PATH = "subscribers.json"
)

func pathOr(envKey, dflt string) string {
	if p := os.Getenv(envKey); p != "" {
		return p
	}
	return dflt
}

func seenPath() string { return pathOr("SEEN_PATH", DEFAULT_SEEN_PATH) }
func subsPath() string { return pathOr("SUBS_PATH", DEFAULT_SUBS_PATH) }

func loadIDs(path string) ids {
	out := ids{}
	data, err := os.ReadFile(path)

	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("read %s: %v", path, err)
		}
		return out
	}

	var ids []int64
	if err := json.Unmarshal(data, &ids); err != nil {
		log.Printf("parse %s: %v", path, err)
		return out
	}
	for _, id := range ids {
		out[id] = true
	}
	return out
}

func saveIDs(path string, m ids) {
	ids := make([]int64, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}

	data, err := json.Marshal(ids)
	if err != nil {
		log.Printf("marshal %s: %v", path, err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("write %s: %v", path, err)
	}
}

func loadSeen() ids         { return loadIDs(seenPath()) }
func saveSeen(m ids)        { saveIDs(seenPath(), m) }
func loadSubscribers() ids  { return loadIDs(subsPath()) }
func saveSubscribers(m ids) { saveIDs(subsPath(), m) }
