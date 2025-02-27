package handlers

import (
	"encoding/json"
	"link/src/utils"
	"log"
	"net/http"
)

// EncodeKeyHandler converts a hex public key to npub and returns it
func EncodeKeyHandler(w http.ResponseWriter, r *http.Request) {
	// Parse publicKey from request
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hexPubKey := r.FormValue("publicKey")
	if hexPubKey == "" {
		http.Error(w, "Missing publicKey", http.StatusBadRequest)
		return
	}

	// Convert to npub
	npub, err := utils.EncodeNpub(hexPubKey)
	if err != nil {
		log.Printf("Failed to encode public key: %v\n", err)
		http.Error(w, "Failed to encode public key", http.StatusInternalServerError)
		return
	}

	// Return npub as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"npub": npub})
}
