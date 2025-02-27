package utils

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"link/src/types"

	"github.com/gorilla/websocket"
)

const WebSocketTimeout = 3 * time.Second // Increased timeout

// FetchUserMetadata fetches the latest kind: 0 profile event from all relays
func FetchUserMetadata(publicKey string, relays []string) (*types.UserMetadata, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var latestEvent *types.NostrEvent
	var latestCreatedAt int64

	for _, url := range relays {
		wg.Add(1)

		go func(relayURL string) {
			defer wg.Done()
			log.Printf("🔍 Connecting to relay: %s\n", relayURL)

			conn, _, err := websocket.DefaultDialer.Dial(relayURL, nil)
			if err != nil {
				log.Printf("❌ WebSocket connection failed (%s): %v\n", relayURL, err)
				return
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
				return
			}

			log.Printf("📡 Sending request to %s: %s\n", relayURL, requestJSON)
			if err := conn.WriteMessage(websocket.TextMessage, requestJSON); err != nil {
				log.Printf("❌ Failed to send request to %s: %v\n", relayURL, err)
				return
			}

			// Wait for response
			conn.SetReadDeadline(time.Now().Add(WebSocketTimeout))

			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("⚠️ Error reading from relay %s: %v\n", relayURL, err)
					return
				}

				var response []interface{}
				if err := json.Unmarshal(message, &response); err != nil {
					log.Printf("❌ Failed to parse response from %s: %v\n", relayURL, err)
					return
				}

				if response[0] == "EVENT" {
					var event types.NostrEvent
					eventData, _ := json.Marshal(response[2])
					if err := json.Unmarshal(eventData, &event); err != nil {
						log.Printf("❌ Failed to parse event JSON from %s: %v\n", relayURL, err)
						continue
					}

					log.Printf("📜 Received event from %s: %+v\n", relayURL, event)

					mu.Lock()
					if event.CreatedAt > latestCreatedAt {
						latestCreatedAt = event.CreatedAt
						latestEvent = &event
					}
					mu.Unlock()
				}
			}
		}(url)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	if latestEvent == nil {
		log.Println("❌ No metadata events received.")
		return nil, nil
	}

	// Parse metadata content
	var metadata types.UserMetadata
	if err := json.Unmarshal([]byte(latestEvent.Content), &metadata); err != nil {
		log.Printf("❌ Failed to parse metadata JSON: %v\n", err)
		return nil, err
	}

	// ✅ Preserve all tags, not just donation ones
	metadata.Tags = latestEvent.Tags

	log.Printf("✅ Latest metadata selected: %+v\n", metadata)
	return &metadata, nil
}
