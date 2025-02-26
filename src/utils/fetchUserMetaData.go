package utils

import (
	"encoding/json"
	"log"
	"time"

	"link/src/types"

	"github.com/gorilla/websocket"
)

const WebSocketTimeout = 2 * time.Second // Set timeout duration

// FetchUserMetadata fetches the latest kind: 0 profile event
func FetchUserMetadata(publicKey string, relays []string) (*types.UserMetadata, error) {
	for _, url := range relays {
		log.Printf("🔍 Connecting to relay: %s\n", url)
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("❌ WebSocket connection failed: %v\n", err)
			continue
		}
		defer conn.Close()

		// Request profile data
		filter := types.SubscriptionFilter{
			Authors: []string{publicKey},
			Kinds:   []int{0}, // Kind 0 = Metadata
		}

		requestJSON, err := json.Marshal([]interface{}{"REQ", "sub1", filter})
		if err != nil {
			log.Printf("❌ Failed to marshal request: %v\n", err)
			return nil, err
		}

		log.Printf("📡 Sending request: %s\n", requestJSON)

		if err := conn.WriteMessage(websocket.TextMessage, requestJSON); err != nil {
			log.Printf("❌ Failed to send request: %v\n", err)
			return nil, err
		}

		// Listen for response
		msgChan := make(chan []byte)
		errChan := make(chan error)

		go func() {
			_, message, err := conn.ReadMessage()
			if err != nil {
				errChan <- err
			} else {
				msgChan <- message
			}
		}()

		select {
		case message := <-msgChan:
			log.Printf("✅ Received WebSocket message: %s\n", message)

			var response []interface{}
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("❌ Failed to parse response: %v\n", err)
				continue
			}

			if response[0] == "EVENT" {
				var event types.NostrEvent
				eventData, _ := json.Marshal(response[2])
				_ = json.Unmarshal(eventData, &event)

				log.Printf("📜 Received event: %+v\n", event)

				// Parse metadata
				var metadata types.UserMetadata
				if err := json.Unmarshal([]byte(event.Content), &metadata); err != nil {
					log.Printf("❌ Failed to parse metadata JSON: %v\n", err)
					continue
				}

				// Extract donation tags
				var donationTags [][]string
				for _, tag := range event.Tags {
					if len(tag) >= 4 && tag[0] == "i" {
						donationTags = append(donationTags, tag)
					}
				}

				log.Printf("✅ Extracted donation tags: %+v\n", donationTags)

				metadata.Tags = donationTags // Store in struct
				return &metadata, nil
			}
		case err := <-errChan:
			log.Printf("❌ WebSocket error: %v\n", err)
		case <-time.After(2 * time.Second):
			log.Println("⏳ WebSocket timeout")
		}
	}
	return nil, nil
}
