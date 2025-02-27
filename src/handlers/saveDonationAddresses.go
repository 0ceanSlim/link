package handlers

import (
	"encoding/json"
	"link/src/utils"
	"log"
	"net/http"

	"github.com/nbd-wtf/go-nostr"
)

// SaveDonationAddresses handles adding/removing donation addresses
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

	// Fetch relays from session
	relays, ok := session.Values["relays"].(utils.RelayList)
	if !ok || len(relays.Both) == 0 {
		http.Error(w, "No relays found", http.StatusInternalServerError)
		return
	}

	// ✅ Send event to Nostr relays & wait for at least one success
	success := false
	for _, relay := range relays.Both {
		ack, err := utils.SendToRelay(relay, signedEvent)
		if err == nil && ack {
			success = true
			break
		}
	}

	if !success {
		http.Error(w, "No relay acknowledged the event", http.StatusBadGateway)
		return
	}

	// ✅ Fetch updated metadata from relays
	updatedMetadata, err := utils.FetchUserMetadata(publicKey, relays.ToStringSlice())
	if err != nil || updatedMetadata == nil {
		log.Println("⚠️ Failed to fetch updated user metadata")
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// ✅ Update session with new donation addresses
	session.Values["donationTags"] = updatedMetadata.Tags
	if err := session.Save(r, w); err != nil {
		log.Printf("❌ Error saving session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// ✅ Return updated donation list to frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"donationTags": updatedMetadata.Tags,
	})
}
