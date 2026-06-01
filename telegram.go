package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	env "github.com/joho/godotenv"
)

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

func broadcast(chatIDs []int64, message string) {
	for _, chatID := range chatIDs {
		if err := sendMessage(chatID, message, "HTML"); err != nil {
			log.Printf("send to %d: %v", chatID, err)
		}
	}
}
