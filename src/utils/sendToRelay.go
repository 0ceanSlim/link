package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nbd-wtf/go-nostr"
)

// SendToRelay sends the signed Nostr event to a relay and ensures acknowledgment
func SendToRelay(relayURL string, event nostr.Event) (bool, error) {
	conn, _, err := websocket.DefaultDialer.Dial(relayURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to connect to relay: %v", err)
	}
	defer conn.Close()

	// Prepare and send the message
	message := []interface{}{"EVENT", event}
	err = conn.WriteJSON(message)
	if err != nil {
		return false, fmt.Errorf("failed to send event to relay: %v", err)
	}

	// **Wait for acknowledgment**
	ackChan := make(chan bool)
	errChan := make(chan error)

	go func() {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second)) // Timeout
		_, reply, err := conn.ReadMessage()
		if err != nil {
			errChan <- err
			return
		}

		// Parse relay response
		var response []interface{}
		if err := json.Unmarshal(reply, &response); err != nil {
			errChan <- err
			return
		}

		if len(response) >= 3 && response[0] == "OK" && response[2] == true {
			log.Printf("✅ Relay %s accepted event", relayURL)
			ackChan <- true
		} else {
			log.Printf("⚠️ Relay %s did not acknowledge event", relayURL)
			ackChan <- false
		}
	}()

	select {
	case ack := <-ackChan:
		return ack, nil
	case err := <-errChan:
		return false, err
	case <-time.After(3 * time.Second): // Timeout if no response
		return false, fmt.Errorf("timeout waiting for relay response")
	}
}
