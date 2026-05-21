package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var subscribers = map[int64]bool{}

func loadToken() string {
	godotenv.Load(".env")
	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatal("BOT_TOKEN not set")
	}
	return token
}

func sendTelegram(chatID int64, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", loadToken())

	body, _ := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func broadcast(message string) {
	for chatID := range subscribers {
		err := sendTelegram(chatID, message)
		if err != nil {
			log.Printf("Failed to send to %d: %v", chatID, err)
		}
	}
}

func pollUpdates(offset int) ([]Update, int) {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30",
		loadToken(), offset,
	)
	resp, err := http.Get(url)

	if err != nil {
		log.Println("Error polling updates:", err)
		return nil, offset
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result UpdateResponse
	json.Unmarshal(data, &result)

	newOffset := offset
	for _, u := range result.Result {
		if u.UpdateID >= newOffset {
			newOffset = u.UpdateID + 1
		}
	}
	return result.Result, newOffset
}

func handleUpdates() {
	offset := 0
	for {
		updates, newOffset := pollUpdates(offset)
		offset = newOffset

		for _, u := range updates {
			if strings.TrimSpace(u.Message.Text) == "/start" {
				chatID := u.Message.Chat.ID

				if !subscribers[chatID] {
					subscribers[chatID] = true
					log.Printf("New subscriber: %d", chatID)

					sendTelegram(
						chatID, "You're subscribed to Tralee apartment alerts!",
					)
				}
			}
		}
	}
}

func main() {
	log.Println("Bot started...")
	go handleUpdates()

	checkListings()
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		checkListings()
	}
}
