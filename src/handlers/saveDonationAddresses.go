package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"link/src/utils"

	"github.com/nbd-wtf/go-nostr"
)

// SaveDonationAddresses saves signed donation addresses and updates session data
func SaveDonationAddresses(w http.ResponseWriter, r *http.Request) {
	session, _ := User.Get(r, "session-name")

	// Extract public key from session
	publicKey, ok := session.Values["publicKey"].(string)
	if !ok || publicKey == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Decode signed event from request body
	var signedEvent nostr.Event
	if err := json.NewDecoder(r.Body).Decode(&signedEvent); err != nil {
		log.Printf("❌ Error decoding signed event: %v", err)
		http.Error(w, "Invalid signed event", http.StatusBadRequest)
		return
	}

	// Get relay list from session
	relays, ok := session.Values["relays"].(utils.RelayList)
	if !ok || len(relays.Both) == 0 {
		http.Error(w, "No relays found", http.StatusInternalServerError)
		return
	}

	// Send event to relays
	results := map[string]string{}
	for _, relay := range relays.Both {
		err := utils.SendToRelay(relay, signedEvent)
		if err != nil {
			log.Printf("⚠️ Failed to send event to relay %s: %v", relay, err)
			results[relay] = fmt.Sprintf("Failed: %v", err)
		} else {
			results[relay] = "Success"
		}
	}

	// Wait briefly to ensure relay processes the event
	time.Sleep(200)

	// Fetch updated user metadata
	updatedMetadata, err := utils.FetchUserMetadata(publicKey, relays.ToStringSlice())
	if err != nil || updatedMetadata == nil {
		log.Println("⚠️ Failed to fetch updated user metadata")
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Update session with new donation tags
	session.Values["donationTags"] = updatedMetadata.Tags
	session.Save(r, w)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
