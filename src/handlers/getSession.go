package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// GetSessionHandler returns the current user's session data as JSON
func GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := User.Get(r, "session-name")

	// Convert session values into a map
	sessionData := map[string]interface{}{
		"UserPublicKey": session.Values["UserPublicKey"],
	}

	// Log session data for debugging
	log.Printf("🔍 Returning session data: %+v", sessionData)

	// Set JSON response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode session data as JSON and send response
	if err := json.NewEncoder(w).Encode(sessionData); err != nil {
		log.Printf("❌ Failed to encode session data: %v", err)
		http.Error(w, "Failed to retrieve session data", http.StatusInternalServerError)
	}
}
