package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"link/src/utils"

	"github.com/nbd-wtf/go-nostr"
)

// SaveDonationAddresses updates donation addresses and ensures session updates immediately
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

	// **Send the signed event to Nostr relays**
	success := false
	results := map[string]string{}
	for _, relay := range relays.Both {
		ack, err := utils.SendToRelay(relay, signedEvent)
		if err != nil {
			log.Printf("⚠️ Failed to send event to relay %s: %v", relay, err)
			results[relay] = fmt.Sprintf("Failed: %v", err)
		} else if ack {
			log.Printf("✅ Relay %s acknowledged event", relay)
			results[relay] = "Success"
			success = true
		} else {
			log.Printf("⚠️ Relay %s did not acknowledge the event", relay)
			results[relay] = "No acknowledgment"
		}
	}

	// **Ensure at least one relay confirmed before proceeding**
	if !success {
		http.Error(w, "No relay acknowledged the event", http.StatusBadGateway)
		return
	}

	// ✅ **Fetch latest metadata (only if event was confirmed)**
	updatedMetadata, err := utils.FetchUserMetadata(publicKey, relays.ToStringSlice())
	if err != nil || updatedMetadata == nil {
		log.Println("⚠️ Failed to fetch updated user metadata")
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// ✅ **Update session with latest `w` tags**
	session.Values["donationTags"] = updatedMetadata.Tags
	log.Printf("✅ Updated Session Donation Tags: %+v", updatedMetadata.Tags)

	if err := session.Save(r, w); err != nil {
		log.Printf("❌ Error saving session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
