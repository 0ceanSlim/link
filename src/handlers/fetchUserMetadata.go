package handlers

import (
	"encoding/json"
	"link/src/utils"
	"net/http"
)

// FetchUserMetadataHandler exposes user metadata to the frontend
func FetchUserMetadataHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := User.Get(r, "session-name")
	publicKey, ok := session.Values["publicKey"].(string)

	if !ok || publicKey == "" {
		sendJSONError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Fetch relays from session
	relays, ok := session.Values["relays"].(utils.RelayList)
	if !ok || len(relays.Both) == 0 {
		sendJSONError(w, "No relays found", http.StatusInternalServerError)
		return
	}

	// Fetch metadata using the utility function
	userMetadata, err := utils.FetchUserMetadata(publicKey, relays.ToStringSlice())
	if err != nil || userMetadata == nil {
		sendJSONError(w, "Failed to fetch user metadata", http.StatusInternalServerError)
		return
	}

	// ✅ Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userMetadata)
}

// ✅ **Helper to send JSON errors instead of HTML**
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
