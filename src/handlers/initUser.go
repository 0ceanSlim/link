package handlers

import (
	"encoding/gob"
	"log"
	"net/http"

	"link/src/utils"

	"github.com/gorilla/sessions"
)

var User = sessions.NewCookieStore([]byte("your-secret-key"))

func init() {
	// Register the RelayList type with gob
	gob.Register(utils.RelayList{})
	gob.Register([][]string{})
}

func InitUser(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler called")

	if err := r.ParseForm(); err != nil {
		log.Printf("Failed to parse form: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	publicKey := r.FormValue("publicKey")
	if publicKey == "" {
		log.Println("Missing publicKey in form data")
		http.Error(w, "Missing publicKey", http.StatusBadRequest)
		return
	}

	log.Printf("Received publicKey: %s\n", publicKey)

	// Fetch user relay list
	initialRelays := []string{
		"wss://purplepag.es", "wss://relay.damus.io", "wss://nos.lol",
		"wss://relay.primal.net", "wss://relay.nostr.band", "wss://offchain.pub",
	}
	userRelays, err := utils.FetchUserRelays(publicKey, initialRelays)
	if err != nil {
		log.Printf("Failed to fetch user relays: %v\n", err)
		http.Error(w, "Failed to fetch user relays", http.StatusInternalServerError)
		return
	}
	log.Printf("Fetched user relays: %+v\n", userRelays)

	// Combine all relays
	allRelays := append(userRelays.Read, userRelays.Write...)
	allRelays = append(allRelays, userRelays.Both...)

	// Fetch user metadata
	userContent, err := utils.FetchUserMetadata(publicKey, allRelays)
	if err != nil {
		log.Printf("Failed to fetch user metadata: %v\n", err)
		http.Error(w, "Failed to fetch user metadata", http.StatusInternalServerError)
		return
	}

	// 🛑 ADD A CHECK FOR NIL METADATA 🛑
	if userContent == nil {
		log.Println("⚠️ Warning: Received nil user metadata!")
		http.Error(w, "No metadata found for user", http.StatusNotFound)
		return
	}

	log.Printf("Fetched user metadata: %+v\n", userContent)

	// Extract only donation-related `w` tags
	var donationTags [][]string
	for _, tag := range userContent.Tags {
		if len(tag) >= 3 && tag[0] == "w" {
			donationTags = append(donationTags, tag) // ["w", "ASSET", "ADDRESS", "NETWORK(optional)"]
		}
	}

	log.Printf("✅ Extracted Donation Tags: %+v", donationTags)

	// Convert publicKey to npub
	npub, err := utils.EncodeNpub(publicKey)
	if err != nil {
		log.Printf("Failed to encode publicKey to npub: %v\n", err)
		http.Error(w, "Failed to encode public key", http.StatusInternalServerError)
		return
	}

	// Save session
	session, _ := User.Get(r, "session-name")
	session.Values["UserPublicKey"] = publicKey // Store logged-in user's public key in the correct field
	session.Values["displayName"] = userContent.DisplayName
	session.Values["picture"] = userContent.Picture
	session.Values["about"] = userContent.About
	session.Values["relays"] = userRelays
	session.Values["donationTags"] = donationTags // ✅ Store donation addresses

	// ✅ Check if session saves successfully
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v\n", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	log.Println("✅ Session saved successfully")

	// Redirect to /p/<npub>
	http.Redirect(w, r, "/"+npub, http.StatusSeeOther)
	log.Printf("🔄 Redirecting to /%s", npub)
}

