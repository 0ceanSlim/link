package handlers

import (
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"

	"link/src/types"
	"link/src/utils"

	"github.com/gorilla/sessions"
)

var User = sessions.NewCookieStore([]byte("your-secret-key"))

func init() {
	// Register types for session storage
	gob.Register(utils.RelayList{})
	gob.Register([][]string{})
	gob.Register(types.NostrEvent{})      // ✅ Fix: Register NostrEvent
	gob.Register([]types.NostrEvent{})    // ✅ Fix: Register slice of NostrEvent
}

// InitUser initializes a user session after login
func InitUser(w http.ResponseWriter, r *http.Request) {
	log.Println("🔑 InitUser called")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Failed to parse form: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	publicKey := r.FormValue("publicKey")
	if publicKey == "" {
		log.Println("❌ Missing publicKey in form data")
		http.Error(w, "Missing publicKey", http.StatusBadRequest)
		return
	}
	log.Printf("✅ Received publicKey: %s", publicKey)

	// Fetch user relay list
	initialRelays := []string{
		"wss://purplepag.es", "wss://relay.damus.io", "wss://nos.lol",
		"wss://relay.primal.net", "wss://relay.nostr.band", "wss://offchain.pub",
	}

	userRelays, err := utils.FetchUserRelays(publicKey, initialRelays)
	if err != nil {
		log.Printf("❌ Failed to fetch user relays: %v", err)
		http.Error(w, "Failed to fetch user relays", http.StatusInternalServerError)
		return
	}

	// Combine relays
	allRelays := append(userRelays.Read, userRelays.Write...)
	allRelays = append(allRelays, userRelays.Both...)
	log.Printf("✅ Fetched user relays: %+v", userRelays)

	// Fetch raw metadata event
	userMetadataEvent, err := utils.FetchUserMetadata(publicKey, allRelays)
	if err != nil {
		log.Printf("❌ Failed to fetch user metadata: %v", err)
		http.Error(w, "Failed to fetch user metadata", http.StatusInternalServerError)
		return
	}
	if userMetadataEvent == nil {
		log.Println("⚠️ No metadata found for user")
		http.Error(w, "No metadata found for user", http.StatusNotFound)
		return
	}

	log.Printf("✅ Fetched raw user metadata event: %+v", userMetadataEvent)

	// Parse metadata content into `UserMetadata`
	var userMetadata types.UserMetadata
	if err := json.Unmarshal([]byte(userMetadataEvent.Content), &userMetadata); err != nil {
		log.Printf("❌ Failed to parse metadata JSON: %v", err)
		http.Error(w, "Failed to parse user metadata", http.StatusInternalServerError)
		return
	}

	// ✅ Store all tags instead of just donation tags
	allTags := userMetadataEvent.Tags
	log.Printf("✅ Extracted All Tags: %+v", allTags)

	// Convert hex public key to npub
	npub, err := utils.EncodeNpub(publicKey)
	if err != nil {
		log.Printf("❌ Failed to encode publicKey to npub: %v", err)
		http.Error(w, "Failed to encode public key", http.StatusInternalServerError)
		return
	}

	// Save session
	session, _ := User.Get(r, "session-name")
	session.Values["UserPublicKey"] = publicKey
	session.Values["displayName"] = userMetadata.DisplayName
	session.Values["picture"] = userMetadata.Picture
	session.Values["about"] = userMetadata.About
	session.Values["relays"] = userRelays
	session.Values["tags"] = userMetadataEvent.Tags // ✅ Store all tags

	// ✅ Store only raw metadata *content* (not the full NostrEvent)
	session.Values["rawUserMetadataContent"] = userMetadataEvent.Content

	if err := session.Save(r, w); err != nil {
		log.Printf("❌ Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	log.Println("✅ Session saved successfully")

	// Redirect to /p/<npub>
	http.Redirect(w, r, "/"+npub, http.StatusSeeOther)
	log.Printf("🔄 Redirecting to /%s", npub)
}
