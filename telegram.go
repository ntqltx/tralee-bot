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

	env "github.com/joho/godotenv"
)

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type UpdateResponse struct {
	Result []Update `json:"result"`
}

var subscribers = loadSubscribers()

func loadToken() string {
	env.Load(".env")
	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatal("BOT_TOKEN not set")
	}
	return token
}

func sendMessage(chatID int64, message, parseMode string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", loadToken())

	payload := map[string]any{
		"chat_id": chatID,
		"text":    message,
	}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func broadcast(message string) {
	for chatID := range subscribers {
		if err := sendMessage(chatID, message, "HTML"); err != nil {
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
			if strings.TrimSpace(u.Message.Text) != "/start" {
				continue
			}
			chatID := u.Message.Chat.ID
			if subscribers[chatID] {
				continue
			}
			subscribers[chatID] = true
			saveSubscribers(subscribers)

			log.Printf("New subscriber: %d", chatID)
			sendMessage(chatID, "You're subscribed to Tralee apartment alerts!", "")
		}
	}
}
