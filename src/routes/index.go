package routes

import (
	"encoding/json"
	"link/src/handlers"
	"link/src/types"
	"link/src/utils"
	"log"
	"net/http"
	"strings"
)

// ProfileHandler serves both the index ("/") and profile pages ("/p/<npub>")
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	session, _ := handlers.User.Get(r, "session-name")
	userPublicKey, _ := session.Values["UserPublicKey"].(string)

	// If user is not logged in and is at "/", redirect to login
	if path == "/" {
		if userPublicKey == "" {
			log.Println("⚠️ No UserPublicKey found in session. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Convert hex public key to npub and redirect to /p/<npub>
		npub, err := utils.EncodeNpub(userPublicKey)
		if err != nil {
			log.Printf("❌ Failed to encode UserPublicKey to npub: %v", err)
			http.Error(w, "Failed to encode public key", http.StatusInternalServerError)
			return
		}

		log.Printf("🔄 Redirecting logged-in user to /p/%s", npub)
		http.Redirect(w, r, "/"+npub, http.StatusSeeOther)
		return
	}

	// If the path starts with "/p/", treat it as a profile view
	if strings.HasPrefix(path, "/") {
		npub := strings.TrimPrefix(path, "/")

		// Convert npub to hex pubkey
		pubKey, err := utils.DecodeNpub(npub)
		if err != nil {
			http.Error(w, "Invalid npub", http.StatusBadRequest)
			return
		}

		// Determine if the logged-in user is viewing their own profile
		isOwnProfile := userPublicKey == pubKey

		// Fetch relays for the public key
		relays, err := utils.FetchUserRelays(pubKey, []string{
			"wss://purplepag.es", "wss://relay.damus.io", "wss://nos.lol",
			"wss://relay.primal.net", "wss://relay.nostr.band", "wss://offchain.pub",
		})
		if err != nil {
			log.Printf("Failed to fetch user relays: %v", err)
		}

		// Get user metadata (raw NostrEvent)
		userEvent, err := utils.FetchUserMetadata(pubKey, relays.ToStringSlice())
		if err != nil {
			http.Error(w, "Failed to fetch user metadata", http.StatusInternalServerError)
			return
		}
		if userEvent == nil {
			http.Error(w, "User metadata not found", http.StatusNotFound)
			return
		}

		// Parse `Content` JSON into `UserMetadata`
		var metadata types.UserMetadata
		if err := json.Unmarshal([]byte(userEvent.Content), &metadata); err != nil {
			log.Printf("❌ Failed to parse user metadata JSON: %v", err)
			http.Error(w, "Failed to parse user metadata", http.StatusInternalServerError)
			return
		}

		// Prepare page data
		data := utils.PageData{
			Title:         metadata.DisplayName + "'s Profile",
			DisplayName:   metadata.DisplayName,
			Picture:       metadata.Picture,
			PublicKey:     pubKey,
			UserPublicKey: userPublicKey,
			About:         metadata.About,
			Relays:        *relays,
			DonationTags:  userEvent.Tags, // ✅ Keep all tags, not just donation tags
			IsOwnProfile:  isOwnProfile,
		}

		// Render profile template
		utils.RenderTemplate(w, data, "index.html", false)
		return
	}

	// If the route does not match known paths, return 404
	http.NotFound(w, r)
}
