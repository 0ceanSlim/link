package handlers

import (
	"encoding/json"
	"link/src/cache"
	"log"
	"net/http"
)

func FetchCurrentKind0FromCache(w http.ResponseWriter, r *http.Request) {
	session, _ := User.Get(r, "session-name")
	publicKey, _ := session.Values["UserPublicKey"].(string)

	if publicKey == "" {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	cachedData, found := cache.GetUserData(publicKey)
	if !found {
		http.Error(w, "No cached data found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"rawUserMetadataContent": cachedData.Metadata,
		"relays":                 cachedData.Relays,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Failed to encode cached data: %v", err)
		http.Error(w, "Failed to retrieve cached data", http.StatusInternalServerError)
	}
}
